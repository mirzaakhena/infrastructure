package database

//func NewSQLiteDefault() (db *gorm.DB) {
//
//	db, err := gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
//	if err != nil {
//		panic(err.Error())
//	}
//
//	return db
//}

// func NewPostgresDefault() (db *gorm.DB) {
//
// 	cfg, err := config.ReadConfig()
// 	if err != nil {
// 		panic(err.Error())
// 	}
//
// 	if cfg.User == "" || cfg.Password == "" || cfg.Database == "" {
// 		panic(fmt.Errorf("user or password ord databaseName is empty"))
// 	}
//
// 	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%v", cfg.Host, cfg.Port, cfg.User, cfg.Database, cfg.Password, cfg.SSLMode)
//
// 	loggerMode := logger.Silent
//
// 	if cfg.LogMode {
// 		loggerMode = logger.Info
// 	}
//
// 	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
// 		Logger: logger.Default.LogMode(loggerMode),
// 	})
// 	if err != nil {
// 		panic(err.Error())
// 	}
//
// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		panic(err.Error())
// 	}
//
// 	sqlDB.SetMaxIdleConns(10)
//
// 	sqlDB.SetMaxOpenConns(10)
//
// 	sqlDB.SetConnMaxLifetime(10 * time.Second)
//
// 	return db
// }

//type contextDBType string
//
//var ContextDBValue contextDBType = "DB"
//
//// ExtractDB is used by other repo to extract the databasex from context
//func ExtractDB(ctx context.Context) (*gorm.DB, error) {
//
//	db, ok := ctx.Value(ContextDBValue).(*gorm.DB)
//	if !ok {
//		return nil, fmt.Errorf("database is not found in context")
//	}
//
//	return db, nil
//}
//
//type GormWithoutTransactionImpl struct {
//	db *gorm.DB
//}
//
//func NewGormWithoutTransactionImpl(db *gorm.DB) *GormWithoutTransactionImpl {
//	return &GormWithoutTransactionImpl{
//		db: db,
//	}
//}
//
//func (r *GormWithoutTransactionImpl) GetDatabase(ctx context.Context) (context.Context, error) {
//	trxCtx := context.WithValue(ctx, ContextDBValue, r.db)
//	return trxCtx, nil
//}
//
//func (r *GormWithoutTransactionImpl) Close(ctx context.Context) error {
//	return nil
//}
//
//type GormWithTransaction struct {
//	db *gorm.DB
//}
//
//func NewGormWithTransaction(db *gorm.DB) *GormWithTransaction {
//	return &GormWithTransaction{
//		db: db,
//	}
//}
//
//func (r *GormWithTransaction) BeginTransaction(ctx context.Context) (context.Context, error) {
//	dbTrx := r.db.Begin()
//	trxCtx := context.WithValue(ctx, ContextDBValue, dbTrx)
//	return trxCtx, nil
//}
//
//func (r *GormWithTransaction) CommitTransaction(ctx context.Context) error {
//	db, err := ExtractDB(ctx)
//	if err != nil {
//		return err
//	}
//	return db.Commit().Error
//}
//
//func (r *GormWithTransaction) RollbackTransaction(ctx context.Context) error {
//	db, err := ExtractDB(ctx)
//	if err != nil {
//		return err
//	}
//	return db.Rollback().Error
//}
