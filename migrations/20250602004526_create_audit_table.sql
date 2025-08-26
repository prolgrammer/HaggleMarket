-- +goose Up
-- +goose StatementBegin
create type user_role as enum ('user', 'shop', 'admin', 'system');
create table audit_logs (
    id bigserial primary key,
    method varchar(10),
    url text,
    resp_status int,
    user_ip text,
    user_id int,
    user_role user_role,
    received_at timestamptz,
    req_body jsonb,
    resp_body jsonb
);

create index audit_logs_user_id_idx on audit_logs (user_id);
create index audit_logs_received_at_idx on audit_logs (received_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table audit_logs;
drop type user_role;
-- +goose StatementEnd
