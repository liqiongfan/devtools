package orm

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbOption struct {
	ip                        string
	port                      int
	username                  string
	password                  string
	charset                   string
	database                  string
	connMaxLifetime           int
	maxIdleCons               int
	connMaxIdleTime           int
	maxOpenCons               int
	slowThreshold             int
	logLevel                  logger.LogLevel
	ignoreRecordNotFoundError bool
	colorful                  bool
}

type DbOption func(*dbOption)

func WithIp(ip string) DbOption {
	return func(o *dbOption) {
		o.ip = ip
	}
}

func WithPort(port int) DbOption {
	return func(o *dbOption) {
		o.port = port
	}
}

func WithUserName(userName string) DbOption {
	return func(o *dbOption) {
		o.username = userName
	}
}

func WithPassword(password string) DbOption {
	return func(o *dbOption) {
		o.password = password
	}
}

func WithCharset(charset string) DbOption {
	return func(o *dbOption) {
		o.charset = charset
	}
}

func WithDatabase(database string) DbOption {
	return func(o *dbOption) {
		o.database = database
	}
}

func WithConnMaxLifetime(connMaxLifetime int) DbOption {
	return func(o *dbOption) {
		o.connMaxLifetime = connMaxLifetime
	}
}

func WithMaxIdleCons(maxIdleCons int) DbOption {
	return func(o *dbOption) {
		o.maxIdleCons = maxIdleCons
	}
}

func WithConnMaxIdleTime(connMaxIdleTime int) DbOption {
	return func(o *dbOption) {
		o.connMaxIdleTime = connMaxIdleTime
	}
}

func WithMaxOpenCons(maxOpenCons int) DbOption {
	return func(o *dbOption) {
		o.maxOpenCons = maxOpenCons
	}
}

func WithSlowThreshold(slowThreshold int) DbOption {
	return func(o *dbOption) {
		o.slowThreshold = slowThreshold
	}
}

func WithLogLevel(logLevel logger.LogLevel) DbOption {
	return func(o *dbOption) {
		o.logLevel = logLevel
	}
}

func WithIgnoreRecordNotFoundError(ignoreRecordNotFoundError bool) DbOption {
	return func(o *dbOption) {
		o.ignoreRecordNotFoundError = ignoreRecordNotFoundError
	}
}

func WithColorful(colorful bool) DbOption {
	return func(o *dbOption) {
		o.colorful = colorful
	}
}

// GetDB Get database gorm connection with config
// caller should capture the panic error
func GetDB(options ...DbOption) *gorm.DB {
	option := &dbOption{
		ip:                        "127.0.0.1",
		port:                      3306,
		username:                  "root",
		password:                  "",
		charset:                   "utf8mb4",
		database:                  "",
		connMaxLifetime:           0,
		maxIdleCons:               0,
		connMaxIdleTime:           0,
		maxOpenCons:               0,
		slowThreshold:             200,
		logLevel:                  logger.Warn,
		ignoreRecordNotFoundError: true,
		colorful:                  false,
	}

	for _, o := range options {
		o(option)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		option.username,
		option.password,
		option.ip,
		option.port,
		option.database,
		option.charset,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Duration(option.slowThreshold) * time.Millisecond,
			LogLevel:                  option.logLevel,
			IgnoreRecordNotFoundError: option.ignoreRecordNotFoundError,
			Colorful:                  option.colorful,
		}),
	})
	if err != nil {
		panic(err)
	}

	iDb, err := db.DB()
	if err != nil {
		panic(err)
	}

	iDb.SetConnMaxLifetime(time.Duration(option.connMaxLifetime) * time.Minute * 30)
	iDb.SetMaxIdleConns(option.maxIdleCons)
	iDb.SetConnMaxIdleTime(time.Duration(option.connMaxIdleTime) * time.Minute)
	iDb.SetMaxOpenConns(option.maxOpenCons)

	return db
}
