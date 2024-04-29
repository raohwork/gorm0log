// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package gorm0log

import (
	"os"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type User struct {
	ID   int
	Name string
}

func initLog() {
	w := zerolog.NewConsoleWriter()
	w.NoColor = true
	w.Out = os.Stdout
	log.Logger = zerolog.New(w).Level(zerolog.WarnLevel)
}

func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		TranslateError: true,
		Logger: &Logger{
			Logger: log.Logger,
		},
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SaveUser(db *gorm.DB, u *User) error {
	return db.Save(u).Error
}

func LoadUser(db *gorm.DB, uid int) (*User, error) {
	var u User
	err := db.Where("id", uid).First(&u).Error
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func Example() {
	initLog()
	db, err := initDB()
	if err != nil {
		log.Error().Err(err).Msg("cannot init db")
		return
	}

	// this line shows sql dump
	if err = SaveUser(db.Debug(), &User{Name: "John Doe"}); err != nil {
		log.Error().Err(err).Msg("cannot insert predefined record")
		return
	}

	// no log message
	u, err := LoadUser(db, 1)
	if err != nil {
		log.Error().Err(err).Msg("cannot load predefined record")
		return
	}
	if u.Name != "John Doe" {
		log.Error().Str("name", u.Name).Msg("unexpected record loaded")
		return
	}

	// record not found error is logged to error level by default
	if u, err = LoadUser(db, 2); err == nil {
		log.Error().Interface("user", u).Msg("unexpected user found for id#2")
		return
	}

	// change log level of record not found message to hide it
	tx := db.Session(&gorm.Session{
		Logger: &Logger{
			Logger: log.Logger,
			Config: Config{
				// logs to Debug level
				ErrorLevel: DebugCommonErr,
				// it's shortcut to
				// ErrorLevel: LogErrorAt(UseDebug, CommonError),
			},
		},
	})
	if u, err = LoadUser(tx, 3); err == nil {
		log.Error().Interface("user", u).Msg("unexpected user found for id#2")
		return
	}

	// change level setting of logger to log record not found error again
	if u, err = LoadUser(tx.Debug(), 4); err == nil {
		log.Error().Interface("user", u).Msg("unexpected user found for id#2")
		return
	}

	// disable record not found log, but dumps sql
	tx = db.Session(&gorm.Session{
		Logger: &Logger{
			Logger: log.Logger.Level(zerolog.DebugLevel),
			Config: Config{ErrorLevel: IgnoreCommonErr},
		},
	})
	if u, err = LoadUser(tx, 5); err == nil {
		log.Error().Interface("user", u).Msg("unexpected user found for id#2")
		return
	}

	// dump sql with source file info
	//
	// example output:
	// <nil> DBG dump sql affected_rows=1 source_file=/path/to/this/example_test.go source_line=53 sql="SELECT * FROM `users` WHERE `id` = 1 ORDER BY `users`.`id` LIMIT 1"
	tx = db.Session(&gorm.Session{
		Logger: &Logger{
			Config: Config{Customize: LogSource("_test.go")},
			// uncomment this to actually logs it
			// Logger: log.Logger.Level(zerolog.DebugLevel),
		},
	})
	if _, err = LoadUser(tx, 1); err != nil {
		log.Error().Err(err).Msg("cannot load predefined record")
		return
	}

	// output:<nil> DBG dump sql affected_rows=1 sql="INSERT INTO `users` (`name`) VALUES (\"John Doe\") RETURNING `id`"
	// <nil> ERR a sql error occurred error="record not found" affected_rows=0 sql="SELECT * FROM `users` WHERE `id` = 2 ORDER BY `users`.`id` LIMIT 1"
	// <nil> DBG a sql error occurred error="record not found" affected_rows=0 sql="SELECT * FROM `users` WHERE `id` = 4 ORDER BY `users`.`id` LIMIT 1"
	// <nil> DBG dump sql affected_rows=0 sql="SELECT * FROM `users` WHERE `id` = 5 ORDER BY `users`.`id` LIMIT 1"
}
