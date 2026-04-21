-- name: RecordLoginSuccess :exec
UPDATE "user"
SET
  failed_attempts = 0,
  locked_until = NULL,
  status = 'active',
  last_login_at = now(),
  updated_at = now()
WHERE
  pk = @user_pk;

-- name: RecordLoginFailure :exec
UPDATE "user"
SET
  failed_attempts = failed_attempts + 1,
  updated_at = now()
WHERE
  pk = @user_pk;

-- name: LockAccount :exec
UPDATE "user"
SET
  status = 'locked',
  locked_until = @locked_until,
  updated_at = now()
WHERE
  pk = @user_pk;

-- name: UpdateUserStatus :exec
UPDATE "user"
SET
  status = @status,
  locked_until = CASE WHEN @status = 'active'::user_status THEN NULL ELSE locked_until END,
  failed_attempts = CASE WHEN @status = 'active'::user_status THEN 0 ELSE failed_attempts END,
  updated_at = now()
WHERE
  pk = @user_pk;
