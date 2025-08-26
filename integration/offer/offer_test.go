package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/EM-Stawberry/Stawberry/pkg/email"

	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	imageName  = "postgres:17.4-alpine"
	dbName     = "db_test"
	dbUser     = "postgres"
	dbPassword = "postgres"
)

func GetContainer() *postgres.PostgresContainer {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx, imageName,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(wait.ForLog(`database system is ready to accept connections`).
			WithOccurrence(2).WithPollInterval(time.Second)),
	)
	if err != nil {
		slog.Error("error starting container", "err", err.Error())
		return nil
	}

	err = pgContainer.Snapshot(context.Background())
	if err != nil {
		slog.Error("error snapshotting container", "err", err)
		return nil
	}

	return pgContainer
}

func GetDB(pgContainer *postgres.PostgresContainer) (*sqlx.DB, error) {
	connString, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		slog.Error(err.Error())
		_ = pgContainer.Terminate(context.Background())
		return nil, err
	}

	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		return nil, err
	}

	_ = goose.SetDialect("postgres")

	err = goose.Up(db.DB, `../../migrations`)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		slog.Error(err.Error())
		return nil, err
	}

	_, err = sqlx.LoadFile(db, `../testdata/offer/sql/populate_test_db.sql`)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return db, nil
}

func mockAuthShopOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       1,
			Name:     "user1",
			Password: "no",
			Email:    "user1email",
			Phone:    "user1phone",
			IsStore:  true,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(1))
		c.Set(helpers.UserIsStoreKey, true)
		c.Set(helpers.UserName, "user1")
		c.Set(helpers.UserEmail, "user1email")
		c.Next()
	}
}

func mockAuthBuyerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       2,
			Name:     "user2",
			Password: "no",
			Email:    "user2email",
			Phone:    "user2phone",
			IsStore:  false,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(2))
		c.Set(helpers.UserIsStoreKey, false)
		c.Set(helpers.UserName, "user2")
		c.Set(helpers.UserEmail, "user2email")
		c.Next()
	}
}

func mockAuthIncorrectShopOwnerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		mockUser := entity.User{
			ID:       3,
			Name:     "user3",
			Password: "no",
			Email:    "user3email",
			Phone:    "user3phone",
			IsStore:  true,
		}
		c.Set("user", mockUser)
		c.Set(helpers.UserIDKey, uint(3))
		c.Set(helpers.UserIsStoreKey, true)
		c.Set(helpers.UserName, "user3")
		c.Set(helpers.UserEmail, "user3email")
		c.Next()
	}
}

type mockMailer struct{}

func newMockMailer() email.MailerService {
	return &mockMailer{}
}

func (m *mockMailer) Registered(userName string, userMail string) {
}

func (m *mockMailer) StatusUpdate(offerID uint, status string, userMail string) {
}

func (m *mockMailer) OfferReceived(offerID uint, userMail string) {
}

func (m *mockMailer) SendGuestOfferNotification(email string, subject string, body string) {
}

func (m *mockMailer) Stop(ctx context.Context) {
}

func setupRouter(authMiddleware gin.HandlerFunc, method, path string, handlerFunc gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.Use(middleware.Errors())
	router.Use(authMiddleware)

	switch method {
	case http.MethodPatch:
		router.PATCH(path, handlerFunc)
	case http.MethodPost:
		router.POST(path, handlerFunc)
	default:
		panic(fmt.Sprintf("unsupported HTTP method: %s", method))
	}
	return router
}

var _ = ginkgo.Describe("offer handlers", ginkgo.Ordered, func() {
	var (
		dbCont    *postgres.PostgresContainer
		db        *sqlx.DB
		offerRepo offer.Repository
		offerServ *offer.Service
		offerHand *handler.OfferHandler
		router    *gin.Engine
	)

	ginkgo.BeforeAll(func() {
		dbCont = GetContainer()
		var err error
		db, err = GetDB(dbCont)
		if err != nil {
			slog.Error(err.Error())
			ginkgo.Fail("Failed to get database connection")
		}
		mailer := newMockMailer()

		offerRepo = repository.NewOfferRepository(db)
		offerServ = offer.NewService(offerRepo, mailer)
		offerHand = handler.NewOfferHandler(offerServ)
	})

	ginkgo.AfterAll(func() {
		_ = db.Close()
		_ = dbCont.Terminate(context.Background())
	})

	ginkgo.Context("when the user is the shop owner", func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthShopOwnerMiddleware(),
				http.MethodPatch, "/api/test/offers/:offerID", offerHand.PatchOfferStatus)
		})

		ginkgo.It("successfully updates the offer status if everything is fine", func() {
			correctOfferID := 1
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

			var ofr dto.PatchOfferStatusResp
			_ = json.Unmarshal(rec.Body.Bytes(), &ofr)
			gomega.Expect(ofr.NewStatus).To(gomega.Equal("accepted"))
		})

		ginkgo.It("fails data validation if the offerID is negative", func() {
			badOfferID := -2
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the offerID is non-numeric", func() {
			badOfferID := "two"
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%s", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the status is not `accepted` or `declined`", func() {
			correctOfferID := 4
			badStatus := "bad_status"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: badStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails data validation if the JSON body is malformed", func() {
			correctOfferID := 4
			malformedJSON := []byte(`{"status": "accepted"`)

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(malformedJSON))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails if the offer is not found", func() {
			badOfferID := 999
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", badOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
		})

		ginkgo.It("fails if the offer is not in a `pending` state", func() {
			correctOfferID := 1 // was changed in the first test to `accepted`
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusConflict))
		})
	})

	ginkgo.Context("when the user is an owner of a different shop", func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthIncorrectShopOwnerMiddleware(),
				http.MethodPatch, "/api/test/offers/:offerID", offerHand.PatchOfferStatus)
		})

		ginkgo.It("fails to update the offer status, even if everything is fine", func() {
			correctOfferID := 2
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusUnauthorized))

		})
	})

	ginkgo.Context("when a user is the creator of an offer", func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthBuyerMiddleware(),
				http.MethodPatch, "/api/test/offers/:offerID", offerHand.PatchOfferStatus)
		})

		ginkgo.It("updates the status to `cancelled` if the request is correct", func() {
			correctOfferID := 2
			correctStatus := "cancelled"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

			var ofr dto.PatchOfferStatusResp
			_ = json.Unmarshal(rec.Body.Bytes(), &ofr)
			gomega.Expect(ofr.NewStatus).To(gomega.Equal("cancelled"))
		})

		ginkgo.It("fails to update the status to `accepted`, "+
			"since that status can only be used by shop owner", func() {
			correctOfferID := 3
			correctStatus := "accepted"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})
	})

	ginkgo.Context("when a user is NOT the creator of an offer", func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthBuyerMiddleware(),
				http.MethodPatch, "/api/test/offers/:offerID", offerHand.PatchOfferStatus)
		})

		ginkgo.It("updates the status to `cancelled` if the request is correct", func() {
			correctOfferID := 5
			correctStatus := "cancelled"
			jsonBody, _ := json.Marshal(struct {
				Status string
			}{
				Status: correctStatus,
			})

			req := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/api/test/offers/%d", correctOfferID),
				bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusNotFound))
		})
	})

	ginkgo.Context("Offer Post Handler for buyers", ginkgo.Ordered, func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthBuyerMiddleware(),
				http.MethodPost, "/api/test/offers", offerHand.PostOffer)
		})

		ginkgo.It("successfully creates an offer with valid data", func() {
			reqBody := dto.PostOfferReq{
				ProductID: 1,
				ShopID:    2,
				Price:     100.50,
				Currency:  "USD",
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/test/offers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusCreated))

			var resp dto.PostOfferResp
			err := json.Unmarshal(rec.Body.Bytes(), &resp)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(resp.ID).To(gomega.BeNumerically(">", 0)) // Expect a positive ID
		})

		ginkgo.It("fails to create an offer with a negative price", func() {
			reqBody := dto.PostOfferReq{
				ProductID: 2,
				ShopID:    2,
				Price:     -10.00,
				Currency:  "USD",
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/test/offers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})

		ginkgo.It("fails to create an offer with missing required fields (e.g., ProductID)", func() {
			reqBody := dto.PostOfferReq{
				// ProductID is missing, which is required by `binding:"required"`
				ShopID:   2,
				Price:    100.00,
				Currency: "USD",
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/test/offers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusBadRequest))
		})
	})

	ginkgo.Context("Offer Post Handler for shop owners", ginkgo.Ordered, func() {
		ginkgo.BeforeEach(func() {
			router = setupRouter(mockAuthIncorrectShopOwnerMiddleware(),
				http.MethodPost, "/api/test/offers", offerHand.PostOffer)
		})

		ginkgo.It("fails to create an offer with a shop owner account", func() {
			reqBody := dto.PostOfferReq{
				ProductID: 4,
				ShopID:    2,
				Price:     100.50,
				Currency:  "USD",
			}
			jsonBody, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/test/offers", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			gomega.Expect(rec.Code).To(gomega.Equal(http.StatusForbidden))
		})
	})
})
