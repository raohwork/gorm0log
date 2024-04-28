# Gorm logger using zerolog as backend

[![Go Reference](https://pkg.go.dev/badge/github.com/raohwork/gorm0log.svg)](https://pkg.go.dev/github.com/raohwork/gorm0log)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/raohwork/gorm0log)
![GitHub License](https://img.shields.io/github/license/raohwork/gorm0log)

# TL; DR

```golang
// setup your logger
w := zerolog.NewConsolewriter()
w.Out = os.Stdout
log.Logger = zerolog.New(w).Level(zerolog.InfoLevel)

// setup gorm
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
	Logger: &gorm0log.Logger{
    	Logger: log.Logger,
        Config: Config{
    		SlowThreshold: 3 * time.Second, // log sql queries slower than 3s
            ErrorLevel: gorm0log.DebugCommonErr, // log common errors at debug level
            ParameterizedQueries: true, // do not log value of parameters
        },
    },
})

db.Debug().Save(&User{Name: "John Doe"}) // dumps sql

var u User
db.Where("id", -1).Find(&user) // record not found, but ignored
db.Debug().Where("id", -1).Find(&user) // record not found error is logged

var x NonExistTable
db.Where("id", 1).Find(&x) // logs error
```

# License

Mozilla Public License, v. 2.0
