-- +goose Up
create table refresh_tokens(
  token text not null primary key,
  created_at timestamp not null,
  updated_at timestamp not null,
  user_id uuid references users(id) on delete cascade not null,
  experies_at timestamp not null,
  revoked_at timestamp
);

-- +goose Down
drop table refresh_tokens;
