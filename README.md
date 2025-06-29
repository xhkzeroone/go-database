# Go Dynamic Repository with GORM

## Giới thiệu
Dự án này cung cấp một dynamic repository pattern cho Go sử dụng GORM, cho phép bạn khai báo các hàm truy vấn động (dynamic query) chỉ bằng cách đặt tên hàm theo cú pháp, tương tự Spring Data JPA. Hỗ trợ nhiều toán tử, phân trang, pool connection, và dễ dàng mở rộng.

## Tính năng nổi bật
- **Dynamic query**: Tự động sinh truy vấn SQL từ tên hàm (FindBy..., FindAllBy..., ...)
- **Hỗ trợ toán tử**: AND, OR, GreaterThan, LessThan, Like, In, Between, IsNull, IsNotNull, OrderBy, Limit
- **Generic repository**: Dùng cho mọi entity/model
- **Cấu hình pool connection**: MaxOpenConns, MaxIdleConns, ConnMaxLifetime
- **Tích hợp GORM, context, transaction**
- **Dễ mở rộng, dễ test (có interface repository)**

## Cài đặt
```bash
go get github.com/xhkzeroone/go-database
```

## Cấu hình database
Cấu hình qua file yaml, env hoặc code (xem file `db/Config.go`):
```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: password
  dbname: postgres
  schema: public
  sslmode: disable
  debug: true
  driver: postgres
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 3600 # giây
```

## Ví dụ sử dụng
```go
type UserModel struct {
    ID        uuid.UUID `gorm:"primarykey;column:id;type:uuid"`
    UserName  string    `gorm:"column:user_name"`
    Status    string    `gorm:"column:status"`
    Total     int       `gorm:"column:total"`
    PartnerId string    `gorm:"column:partner_id"`
    CreatedAt time.Time `gorm:"column:created_at"`
    DeletedAt *time.Time `gorm:"column:deleted_at"`
}

type UserRepository struct {
    *repo.Repository[UserModel, uuid.UUID]
    FindByUserName func(ctx context.Context, username string) (*UserModel, error) `repo:"@Query"`
    FindAllByStatusAndTotalGreaterThanOrderByCreatedAtDescLimit10 func(ctx context.Context, status string, total int) ([]UserModel, error) `repo:"@Query"`
    FindByStatusIn func(ctx context.Context, statuses []string) (*UserModel, error) `repo:"@Query"`
    FindByCreatedAtBetween func(ctx context.Context, from, to time.Time) (*UserModel, error) `repo:"@Query"`
    FindByDeletedAtIsNull func(ctx context.Context) (*UserModel, error) `repo:"@Query"`
}

// Khởi tạo repository và inject hàm động
db, _ := db.Open(cfg)
repo := repo.NewRepository[UserModel, uuid.UUID](db)
r := &UserRepository{Repository: repo}
r.Repository.FillFuncFields(r)

// Gọi hàm động
user, err := r.FindByUserName(ctx, "john")
users, err := r.FindAllByStatusAndTotalGreaterThanOrderByCreatedAtDescLimit10(ctx, "active", 100)
```

## Cú pháp đặt tên hàm dynamic
- **FindBy...And...Or...**: Điều kiện WHERE (AND/OR)
- **OrderBy...Asc/Desc**: Sắp xếp
- **LimitN**: Giới hạn số bản ghi
- **Toán tử**:
  - `GreaterThan`, `LessThan`, `GreaterThanEqual`, `LessThanEqual`, `NotEqual`, `Like`, `In`, `Between`, `IsNull`, `IsNotNull`

### Ví dụ tên hàm:
- `FindByUserNameAndStatus`
- `FindByTotalGreaterThan`
- `FindByStatusIn`
- `FindByCreatedAtBetween`
- `FindByDeletedAtIsNull`

## Cấu hình pool connection
- `max_open_conns`: Số connection tối đa
- `max_idle_conns`: Số connection idle tối đa
- `conn_max_lifetime`: Thời gian sống tối đa của connection (giây)

## Mở rộng
- Bổ sung toán tử mới chỉ cần thêm vào hàm parseMethodName trong `repo/DynamicProxy.go`
- Có thể mở rộng cho các driver khác (sqlite, mssql, ...)
- Dễ dàng mock/test repository qua interface

## Đóng góp
PR, issue, góp ý đều rất hoan nghênh!

---
**Made with ❤️ by xhkzeroone**
