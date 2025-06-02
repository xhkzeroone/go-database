package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/xhkzeroone/go-database/db"
	"github.com/xhkzeroone/go-database/repo"
	"gorm.io/gorm"
	"time"
)

// --- Ví dụ sử dụng:

type UserModel struct {
	ID        uuid.UUID `gorm:"primarykey;column:id;type:uuid"`
	PartnerId string    `gorm:"column:partner_id"`
	Total     int       `gorm:"column:total"`
	UserName  string    `gorm:"column:user_name"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	Email     string    `gorm:"column:email"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (u *UserModel) TableName() string {
	return "user_tbl"
}

func (u *UserModel) BeforeCreate(ctx *gorm.DB) (err error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	return
}

func (u *UserModel) BeforeUpdate(ctx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

func (u *UserModel) GetTotal() int {
	return u.Total
}

type UserRepository struct {
	*repo.Repository[UserModel, uuid.UUID]
	FindByUserName                     func(ctx context.Context, username string) (*UserModel, error)                                 `repo:"@Query"`
	FindByUserNameAndEmailOrPartnerId  func(ctx context.Context, username string, email string, partnerId string) (*UserModel, error) `repo:"@Query"`
	FindAllByEmailOrderByIDDescLimit10 func(ctx context.Context, email string) ([]UserModel, error)                                   `repo:"@Query"`
}

func main() {
	// Giả lập DataSource (bạn thay thành *gorm.DataSource thật)

	//// Dùng mặc định
	//db.Open(cfg)
	//
	//// Custom DSNBuilder cho Oracle
	//type OracleBuilder struct{}
	//func (o *OracleBuilder) Build(cfg *Config) (gorm.Dialector, error) {
	//	return oracle.Open("oracle://..."), nil
	//}
	//db.Open(cfg, db.WithDSNBuilder(&OracleBuilder{}))

	var datab, _ = db.Open(&db.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "keycloak",
		Password: "password",
		DBName:   "keycloak",
		Schema:   "public",
		Debug:    true,
		SSLMode:  "disable",
		Driver:   "postgres",
	})

	repository := repo.NewRepository[UserModel, uuid.UUID](datab)
	r := &UserRepository{
		Repository: repository,
	}

	err := r.Repository.FillFuncFields(r)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	user3, err := repository.FindByID(ctx, uuid.MustParse("78c83478-5e15-4720-9acb-b70ab32f011b"))
	fmt.Println(user3, err)

	_, err = repository.Exists(ctx, "user_name=?", "123")
	fmt.Println(err)

	_, err = repository.CountBy(ctx, "user_name=?", "123")
	fmt.Println(err)
	// Bây giờ bạn có thể gọi
	user, err := r.FindByUserName(ctx, "123")
	fmt.Println(user, err)
	//user1, err := r.FindByUserNameAndEmailOrPartnerId(ctx, "123", "test@example.com", "asd")
	//fmt.Println(user1, err)
	//users, err := r.FindAllByEmailOrderByIDDescLimit10(ctx, "test@example.com")
	//fmt.Println(users, err)

	// (GormDB chưa khởi tạo nên ví dụ này chỉ minh họa)
	fmt.Println("Repository methods injected successfully.")
}
