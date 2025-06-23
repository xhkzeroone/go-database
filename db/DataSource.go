package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Option func(o *options)

type options struct {
	dialector  gorm.Dialector
	gormConfig *gorm.Config
	debug      *bool
	dsnBuilder DSNBuilder
}

// DataSource defines common database operations.
type DataSource struct {
	*gorm.DB
}

// DSNBuilder defines how to build a gorm.Dialector based on config.
type DSNBuilder interface {
	Build(*Config) (gorm.Dialector, error)
}

type DefaultDSNBuilder struct{}

func (d *DefaultDSNBuilder) Build(c *Config) (gorm.Dialector, error) {
	switch c.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.DBName, c.Schema, c.SSLMode,
		)
		return postgres.Open(dsn), nil
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.DBName,
		)
		return mysql.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", c.Driver)
	}
}

func WithDSNBuilder(builder DSNBuilder) Option {
	return func(o *options) {
		o.dsnBuilder = builder
	}
}

func WithDialector(d gorm.Dialector) Option {
	return func(o *options) {
		o.dialector = d
	}
}

func WithGormConfig(cfg *gorm.Config) Option {
	return func(o *options) {
		o.gormConfig = cfg
	}
}

func WithDebug(debug bool) Option {
	return func(o *options) {
		o.debug = &debug
	}
}

// Open mở kết nối tới database với config và các option
func Open(cfg *Config, opts ...Option) (*DataSource, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	var dialector gorm.Dialector
	if opt.dialector != nil {
		dialector = opt.dialector
	} else {
		builder := opt.dsnBuilder
		if builder == nil {
			builder = &DefaultDSNBuilder{}
		}
		var err error
		dialector, err = builder.Build(cfg)
		if err != nil {
			log.Printf("failed to build DSN: %v", err)
			return nil, err
		}
	}

	gormCfg := &gorm.Config{}
	if opt.gormConfig != nil {
		gormCfg = opt.gormConfig
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		log.Printf("failed to connect database: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	// Pool config (nếu có trong Config)
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	debugMode := cfg.Debug
	if opt.debug != nil {
		debugMode = *opt.debug
	}
	if debugMode {
		db = db.Debug()
		log.Println("GORM debug mode is enabled")
	}

	log.Println("Successfully connected to database")
	return &DataSource{DB: db}, nil
}

// Close đóng kết nối database
func (p *DataSource) Close() error {
	if p == nil || p.DB == nil {
		return nil
	}
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DefaultDSNBuilder hỗ trợ mở rộng driver (ví dụ: sqlite)
// case "sqlite":
// 	return sqlite.Open(c.DBName), nil
// ... existing code ...
