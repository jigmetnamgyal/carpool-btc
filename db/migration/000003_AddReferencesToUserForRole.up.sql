ALTER TABLE IF EXISTS users
ADD COLUMN role_id bigint,
ADD CONSTRAINT fk_user_roles FOREIGN KEY (role_id) REFERENCES roles (id)