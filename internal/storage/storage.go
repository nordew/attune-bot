package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	userTableName          = "users"
	userSettingsTableName  = "user_settings"
	dayRecordsTableName    = "day_records"
	focusSessionsTableName = "focus_sessions"

	codeUnique = "23505"
)

type Storages struct {
	User         UserStorage
	UserSettings UserSettingsStorage
	DayRecord    DayRecordStorage
	FocusSession FocusSessionStorage
}

func NewStorages(pool *pgxpool.Pool) Storages {
	return Storages{
		User:         NewUserStorage(pool),
		UserSettings: NewUserSettingsStorage(pool),
		DayRecord:    NewDayRecordStorage(pool),
		FocusSession: NewFocusSessionStorage(pool),
	}
}
