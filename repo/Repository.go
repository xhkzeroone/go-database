package repo

import (
	"context"
	"fmt"
	"github.com/xhkzeroone/go-database/db"
	"gorm.io/gorm/schema"
	"reflect"
)

type Page[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"totalCount"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
}

type Repository[T any, ID comparable] struct {
	*db.DataSource
}

func NewRepository[T any, ID comparable](db *db.DataSource) *Repository[T, ID] {
	return &Repository[T, ID]{
		DataSource: db,
	}
}

func (r *Repository[T, ID]) Insert(ctx context.Context, entity *T) error {
	return r.WithContext(ctx).Model(new(T)).Create(entity).Error
}

func (r *Repository[T, ID]) FindByID(ctx context.Context, id ID) (*T, error) {
	entity := new(T)
	err := r.WithContext(ctx).Model(new(T)).First(entity, id).Error
	return entity, err
}

func (r *Repository[T, ID]) Select(ctx context.Context, query any, args ...any) ([]T, error) {
	var list []T
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Find(&list).Error
	return list, err
}

func (r *Repository[T, ID]) SelectOne(ctx context.Context, query any, args ...any) (*T, error) {
	var item T
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).First(&item).Error
	return &item, err
}

func (r *Repository[T, ID]) Update(ctx context.Context, entity *T) error {
	return r.WithContext(ctx).Model(new(T)).Save(entity).Error
}

func (r *Repository[T, ID]) DeleteByID(ctx context.Context, id ID) error {
	return r.WithContext(ctx).Model(new(T)).Delete(new(T), id).Error
}

func (r *Repository[T, ID]) ListAll(ctx context.Context) ([]T, error) {
	var list []T
	err := r.WithContext(ctx).Model(new(T)).Find(&list).Error
	return list, err
}

func (r *Repository[T, ID]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.WithContext(ctx).Model(new(T)).Count(&count).Error
	return count, err
}

func (r *Repository[T, ID]) CountBy(ctx context.Context, query any, args ...any) (int64, error) {
	var count int64
	err := r.WithContext(ctx).Model(new(T)).Where(query, args...).Count(&count).Error
	return count, err
}

func (r *Repository[T, ID]) RawQuery(ctx context.Context, query string, args ...any) ([]T, error) {
	var results []T
	err := r.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	return results, err
}

func (r *Repository[T, ID]) Exists(ctx context.Context, query any, args ...any) (bool, error) {
	var exists bool
	tableName := r.tableName()
	rawQuery := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", tableName, query)
	err := r.WithContext(ctx).Raw(rawQuery, args...).Scan(&exists).Error
	return exists, err
}

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

func (r *Repository[T, ID]) tableName() string {
	entity := new(T)
	if tn, ok := any(entity).(schema.Tabler); ok {
		return tn.TableName()
	}
	return r.NamingStrategy.TableName(reflect.TypeOf(*entity).Name())
}
