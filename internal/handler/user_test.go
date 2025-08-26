package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/handler/dto"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("UserHandler", func() {
	var (
		ctrl        *gomock.Controller
		mockService *MockUserService
		cfg         *config.Config
		handler     *UserHandler
		router      *gin.Engine
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockService = NewMockUserService(ctrl)
		cfg = &config.Config{}
		handler = NewUserHandler(cfg, mockService)

		gin.SetMode(gin.TestMode)
		router = gin.New()
		router.Use(middleware.Errors())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Registration", func() {
		BeforeEach(func() {
			router.POST("/register", handler.Registration)
		})

		Context("when registration is successful", func() {
			It("should return access and refresh tokens", func() {
				input := dto.RegistrationUserReq{
					Name:        "Test User",
					Email:       "test@example.com",
					Password:    "password123",
					Phone:       "1234567890",
					Fingerprint: "fp123",
				}

				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), "fp123").
					Return("access_token", "refresh_token", nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response dto.RegistrationUserResp
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.AccessToken).To(Equal("access_token"))
				Expect(response.RefreshToken).To(Equal("refresh_token"))
			})
		})

		Context("when request has missing required fields", func() {
			It("should return bad request", func() {
				input := dto.RegistrationUserReq{
					Email:       "test@example.com",
					Password:    "password123",
					Fingerprint: "fp123",
				}

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when service returns an error", func() {
			It("should return internal server error", func() {
				input := dto.RegistrationUserReq{
					Name:        "Test User",
					Email:       "test@example.com",
					Password:    "password123",
					Phone:       "1234567890",
					Fingerprint: "fp123",
				}

				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), "fp123").
					Return("", "", apperror.New(apperror.InternalError, "service error", nil))

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when JSON is invalid", func() {
			It("should return bad request", func() {
				jsonData := []byte(`{"invalid json"`)
				req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("Login", func() {
		BeforeEach(func() {
			router.POST("/login", handler.Login)
		})

		Context("when login is successful", func() {
			It("should return access and refresh tokens", func() {
				input := dto.LoginUserReq{
					Email:       "test@example.com",
					Password:    "password123",
					Fingerprint: "fp123",
				}

				mockService.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "password123", "fp123").
					Return("access_token", "refresh_token", nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response dto.LoginUserResp
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.AccessToken).To(Equal("access_token"))
				Expect(response.RefreshToken).To(Equal("refresh_token"))
			})
		})

		Context("when authentication fails", func() {
			It("should return unauthorized", func() {
				input := dto.LoginUserReq{
					Email:       "test@example.com",
					Password:    "wrong_password",
					Fingerprint: "fp123",
				}

				mockService.EXPECT().
					Authenticate(gomock.Any(), "test@example.com", "wrong_password", "fp123").
					Return("", "", apperror.ErrIncorrectPassword)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when JSON is invalid", func() {
			It("should return bad request", func() {
				jsonData := []byte(`{"invalid json"`)
				req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("Refresh", func() {
		BeforeEach(func() {
			router.POST("/refresh", handler.Refresh)
		})

		Context("when refresh is successful", func() {
			It("should return new access and refresh tokens", func() {
				input := dto.RefreshReq{
					RefreshToken: "old_refresh_token",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Refresh(gomock.Any(), "old_refresh_token", "fp123").
					Return("new_access_token", "new_refresh_token", nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response dto.RefreshResp
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.AccessToken).To(Equal("new_access_token"))
				Expect(response.RefreshToken).To(Equal("new_refresh_token"))
			})
		})

		Context("when token is empty but cookie is provided", func() {
			It("should use cookie token and return new tokens", func() {
				input := dto.RefreshReq{
					RefreshToken: "",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Refresh(gomock.Any(), "cookie_refresh_token", "fp123").
					Return("new_access_token", "new_refresh_token", nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "cookie_refresh_token",
				})

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))

				var response dto.RefreshResp
				err := json.Unmarshal(w.Body.Bytes(), &response)
				Expect(err).ToNot(HaveOccurred())
				Expect(response.AccessToken).To(Equal("new_access_token"))
				Expect(response.RefreshToken).To(Equal("new_refresh_token"))
			})
		})

		Context("when token is empty and no cookie is provided", func() {
			It("should return bad request", func() {
				input := dto.RefreshReq{
					RefreshToken: "",
					Fingerprint:  "fp123",
				}

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when service returns an error", func() {
			It("should return internal server error", func() {
				input := dto.RefreshReq{
					RefreshToken: "invalid_token",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Refresh(gomock.Any(), "invalid_token", "fp123").
					Return("", "", apperror.New(apperror.InternalError, "invalid refresh token", nil))

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when JSON is invalid", func() {
			It("should return bad request", func() {
				jsonData := []byte(`{"invalid json"`)
				req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("Logout", func() {
		BeforeEach(func() {
			router.POST("/logout", handler.Logout)
		})

		Context("when logout is successful", func() {
			It("should return OK status", func() {
				input := dto.LogoutReq{
					RefreshToken: "refresh_token",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Logout(gomock.Any(), "refresh_token", "fp123").
					Return(nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when token is empty but cookie is provided", func() {
			It("should use cookie token and logout successfully", func() {
				input := dto.LogoutReq{
					RefreshToken: "",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Logout(gomock.Any(), "cookie_refresh_token", "fp123").
					Return(nil)

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "cookie_refresh_token",
				})

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})

		Context("when token is empty and no cookie is provided", func() {
			It("should return bad request", func() {
				input := dto.LogoutReq{
					RefreshToken: "",
					Fingerprint:  "fp123",
				}

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when logout fails", func() {
			It("should return internal server error", func() {
				input := dto.LogoutReq{
					RefreshToken: "invalid_token",
					Fingerprint:  "fp123",
				}

				mockService.EXPECT().
					Logout(gomock.Any(), "invalid_token", "fp123").
					Return(apperror.New(apperror.InternalError, "logout failed", nil))

				jsonData, _ := json.Marshal(input)
				req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when JSON is invalid", func() {
			It("should return bad request", func() {
				jsonData := []byte(`{"invalid json"`)
				req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

func TestUserHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UserHandler Suite")
}
