package auth

import (
	"time"

	"github.com/rizesql/mithras/pkg/db"
)

func checkUserStatus(userStatus db.UserStatus, userLockedUntil *time.Time, now time.Time) error {
	if userStatus == db.UserStatusSuspended {
		return errAccountSuspended
	}

	if userStatus == db.UserStatusLocked && userLockedUntil != nil && userLockedUntil.After(now) {
		return errAccountLocked(userLockedUntil.String())
	}

	return nil
}
