package repo

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// QueryParts Struct query parts
type QueryParts struct {
	WhereClauses []string
	OrderBy      string
	Limit        int
}

// toSnakeCase chuẩn hơn (ví dụ: UserName -> user_name, URLString -> url_string)
func toSnakeCase(s string) string {
	var result []rune
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func parseOrderBy(s string) (field string, direction string) {
	if strings.HasSuffix(s, "Desc") {
		return s[:len(s)-4], "DESC"
	}
	if strings.HasSuffix(s, "Asc") {
		return s[:len(s)-3], "ASC"
	}
	return s, "ASC"
}

// parse method name thành các phần where, orderby, limit
func parseMethodName(methodName string) (*QueryParts, error) {
	const prefix = "FindBy"
	if !strings.HasPrefix(methodName, prefix) {
		return nil, fmt.Errorf("method name must start with %s", prefix)
	}
	methodName = methodName[len(prefix):]

	qp := &QueryParts{}

	// Tách OrderBy
	orderByIdx := strings.Index(methodName, "OrderBy")
	orderByPart := ""
	if orderByIdx >= 0 {
		orderByPart = methodName[orderByIdx+len("OrderBy"):]
		methodName = methodName[:orderByIdx]
	}

	// Tách Limit trong phần orderBy
	limitPart := ""
	limitIdx := strings.Index(orderByPart, "Limit")
	if limitIdx >= 0 {
		limitPart = orderByPart[limitIdx+len("Limit"):]
		orderByPart = orderByPart[:limitIdx]
	}

	// Parse OrderBy
	if orderByPart != "" {
		field, dir := parseOrderBy(orderByPart)
		qp.OrderBy = fmt.Sprintf("%s %s", toSnakeCase(field), dir)
	}

	// Parse Limit
	if limitPart != "" {
		n, err := strconv.Atoi(limitPart)
		if err != nil {
			return nil, fmt.Errorf("invalid limit number: %v", err)
		}
		qp.Limit = n
	}

	// Parse điều kiện WHERE: xử lý AND, OR
	// Tách OR trước
	orParts := strings.Split(methodName, "Or")
	whereClauses := make([]string, 0, len(orParts))

	for _, orPart := range orParts {
		andParts := strings.Split(orPart, "And")
		andClauses := make([]string, 0, len(andParts))
		for _, andPart := range andParts {
			andClauses = append(andClauses, fmt.Sprintf("%s = ?", toSnakeCase(andPart)))
		}
		group := "(" + strings.Join(andClauses, " AND ") + ")"
		whereClauses = append(whereClauses, group)
	}

	qp.WhereClauses = whereClauses

	return qp, nil
}

func buildGormQuery(db *gorm.DB, qp *QueryParts, args []interface{}) *gorm.DB {
	whereClause := strings.Join(qp.WhereClauses, " OR ")
	q := db.Where(whereClause, args...)
	if qp.OrderBy != "" {
		q = q.Order(qp.OrderBy)
	}
	if qp.Limit > 0 {
		q = q.Limit(qp.Limit)
	}
	return q
}

// FillFuncFields inject các func dynamic vào struct repo có tag `repo:"@Query"`
func (r *Repository[T, ID]) FillFuncFields(repo interface{}) error {
	v := reflect.ValueOf(repo).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jpaTag := field.Tag.Get("repo")

		if jpaTag != "@Query" || field.Type.Kind() != reflect.Func {
			continue
		}
		funcType := field.Type
		if funcType.NumIn() == 0 || funcType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
			return fmt.Errorf("method %s must have context.Context as the first parameter", field.Name)
		}

		if jpaTag == "@Query" && field.Type.Kind() == reflect.Func {
			methodName := field.Name

			var qp *QueryParts
			var err error
			var isFindAll bool

			if strings.HasPrefix(methodName, "FindAllBy") {
				isFindAll = true
				qp, err = parseMethodName("FindBy" + methodName[len("FindAllBy"):])
			} else if strings.HasPrefix(methodName, "FindBy") {
				isFindAll = false
				qp, err = parseMethodName(methodName)
			} else {
				return fmt.Errorf("method name %s phải bắt đầu FindBy hoặc FindAllBy", methodName)
			}
			if err != nil {
				return err
			}

			funcType := field.Type

			fn := reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
				ctxVal := args[0]
				ctx := ctxVal.Interface().(context.Context)

				params := make([]interface{}, len(args)-1)
				for i := 1; i < len(args); i++ {
					params[i-1] = args[i].Interface()
				}

				dbWithCtx := r.DataSource.WithContext(ctx).Model(new(T))

				if isFindAll {
					var res []T
					err := buildGormQuery(dbWithCtx, qp, params).Find(&res).Error
					if err != nil {
						results = []reflect.Value{
							reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*T)(nil)).Elem()), 0, 0),
							reflect.ValueOf(err),
						}
						return results
					}

					results = []reflect.Value{
						reflect.ValueOf(res),
						reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()), // nil
					}
				} else {
					var res T
					err := buildGormQuery(dbWithCtx, qp, params).First(&res).Error
					if err != nil {
						results = []reflect.Value{
							reflect.Zero(reflect.TypeOf(&res)),
							reflect.ValueOf(err),
						}
						return results
					}

					results = []reflect.Value{
						reflect.ValueOf(&res),
						reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()),
					}
				}

				return
			})

			v.Field(i).Set(fn)
		}
	}
	return nil
}
