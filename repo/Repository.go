package repo

import (
	"context"
	"errors"
	"reflect"

	"github.com/xhkzeroone/go-database/db"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Page[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"totalCount"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
}

// IRepository định nghĩa interface cho repository generic
// Giúp dễ mock/test trong unit test
// Có thể mở rộng thêm các hàm khác nếu cần
type IRepository[T any, ID comparable] interface {
	Insert(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id ID) (*T, error)
	FindWhere(ctx context.Context, query any, args ...any) ([]T, error)
	FindOneWhere(ctx context.Context, query any, args ...any) (*T, error)
	Update(ctx context.Context, entity *T) error
	DeleteByID(ctx context.Context, id ID) error
	ListAll(ctx context.Context) ([]T, error)
	Count(ctx context.Context) (int64, error)
	CountBy(ctx context.Context, query any, args ...any) (int64, error)
	RawQuery(ctx context.Context, query string, args ...any) ([]T, error)
	Exists(ctx context.Context, query any, args ...any) (bool, error)
	Pageable(ctx context.Context, page int, pageSize int, query any, args ...any) (*Page[T], error)
}

// Repository là struct generic cho thao tác DB với GORM
// T là kiểu entity, ID là kiểu khóa chính
type Repository[T any, ID comparable] struct {
	*db.DataSource
}

// NewRepository khởi tạo repository mới
func NewRepository[T any, ID comparable](db *db.DataSource) *Repository[T, ID] {
	return &Repository[T, ID]{
		DataSource: db,
	}
}

// Insert thêm entity vào DB
func (r *Repository[T, ID]) Insert(ctx context.Context, entity *T) error {
	return r.WithContext(ctx).Model(new(T)).Create(entity).Error
}

// FindByID tìm entity theo ID, trả về nil nếu không tìm thấy
func (r *Repository[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	entity := new(T)
	err := r.WithContext(ctx).Model(new(T)).First(entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return entity, nil
}

// FindWhere tìm danh sách entity theo điều kiện
func (r *Repository[T, ID]) FindWhere(ctx context.Context, query any, args ...any) ([]T, error) {
	var list []T
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Find(&list).Error
	return list, err
}

// FindOneWhere tìm một entity theo điều kiện
func (r *Repository[T, ID]) FindOneWhere(ctx context.Context, query any, args ...any) (*T, error) {
	var item T
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &item, nil
}

// Update cập nhật entity
func (r *Repository[T, ID]) Update(ctx context.Context, entity *T) error {
	return r.WithContext(ctx).Model(new(T)).Save(entity).Error
}

// DeleteByID xóa entity theo ID
func (r *Repository[T, ID]) DeleteByID(ctx context.Context, id ID) error {
	return r.WithContext(ctx).Model(new(T)).Delete(new(T), id).Error
}

// ListAll lấy tất cả entity
func (r *Repository[T, ID]) ListAll(ctx context.Context) ([]T, error) {
	var list []T
	err := r.WithContext(ctx).Model(new(T)).Find(&list).Error
	return list, err
}

// Count đếm tổng số entity
func (r *Repository[T, ID]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.WithContext(ctx).Model(new(T)).Count(&count).Error
	return count, err
}

// CountBy đếm entity theo điều kiện
func (r *Repository[T, ID]) CountBy(ctx context.Context, query any, args ...any) (int64, error) {
	var count int64
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count, err
}

// RawQuery thực thi truy vấn SQL thô
func (r *Repository[T, ID]) RawQuery(ctx context.Context, query string, args ...any) ([]T, error) {
	var results []T
	err := r.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	return results, err
}

// Exists kiểm tra có entity nào thỏa điều kiện không (an toàn, không dùng raw SQL)
func (r *Repository[T, ID]) Exists(ctx context.Context, query any, args ...any) (bool, error) {
	var count int64
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count > 0, err
}

// Pageable phân trang kết quả truy vấn
func (r *Repository[T, ID]) Pageable(ctx context.Context, page int, pageSize int, query any, args ...any) (*Page[T], error) {
	var items []T
	var total int64

	// Đếm tổng số bản ghi
	if err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&total).Error; err != nil {
		return nil, err
	}

	// Lấy dữ liệu theo trang
	offset := (page - 1) * pageSize
	if err := r.WithContext(ctx).Where(query, args...).Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &Page[T]{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// tableName trả về tên bảng của entity
func (r *Repository[T, ID]) tableName() string {
	entity := new(T)
	if tn, ok := any(entity).(schema.Tabler); ok {
		return tn.TableName()
	}
	return r.NamingStrategy.TableName(reflect.TypeOf(*entity).Name())
}
