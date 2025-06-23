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

// extractSuffixPart tách phần hậu tố (ví dụ: OrderBy, Limit)
func extractSuffixPart(s, key string) (before, part string) {
	idx := strings.Index(s, key)
	if idx >= 0 {
		return s[:idx], s[idx+len(key):]
	}
	return s, ""
}

// parseWhereConditions tách các điều kiện WHERE (AND/OR, toán tử)
func parseWhereConditions(methodName string, parseFieldOp func(string) (string, string)) ([]string, []int, error) {
	orConditions := strings.Split(methodName, "Or")
	whereClauses := make([]string, 0, len(orConditions))
	paramCounts := make([]int, 0, len(orConditions))

	for _, orCond := range orConditions {
		andConditions := strings.Split(orCond, "And")
		andClauses := make([]string, 0, len(andConditions))
		countParams := 0
		for _, andCond := range andConditions {
			field, op := parseFieldOp(andCond)
			switch op {
			case "IN":
				andClauses = append(andClauses, fmt.Sprintf("%s IN (?)", toSnakeCase(field)))
				countParams++ // IN nhận 1 tham số là slice
			case "BETWEEN":
				andClauses = append(andClauses, fmt.Sprintf("%s BETWEEN ? AND ?", toSnakeCase(field)))
				countParams += 2 // BETWEEN nhận 2 tham số
			case "IS NULL", "IS NOT NULL":
				andClauses = append(andClauses, fmt.Sprintf("%s %s", toSnakeCase(field), op))
				// IS NULL không nhận tham số
			default:
				andClauses = append(andClauses, fmt.Sprintf("%s %s ?", toSnakeCase(field), op))
				countParams++
			}
		}
		group := "(" + strings.Join(andClauses, " AND ") + ")"
		whereClauses = append(whereClauses, group)
		paramCounts = append(paramCounts, countParams)
	}
	return whereClauses, paramCounts, nil
}

func parseMethodName(rawMethodName string) (*QueryParts, error) {
	const prefix = "FindBy"
	if !strings.HasPrefix(rawMethodName, prefix) {
		return nil, fmt.Errorf("method name must start with %s", prefix)
	}
	methodName := rawMethodName[len(prefix):]

	qp := &QueryParts{}

	// Map các hậu tố sang toán tử SQL
	operatorMap := []struct {
		Suffix string
		SQLOp  string
	}{
		{"GreaterThanEqual", ">="},
		{"LessThanEqual", "<="},
		{"GreaterThan", ">"},
		{"LessThan", "<"},
		{"NotEqual", "!="},
		{"Like", "LIKE"},
		{"In", "IN"},
		{"Between", "BETWEEN"},
		{"IsNull", "IS NULL"},
		{"IsNotNull", "IS NOT NULL"},
	}

	parseFieldOp := func(part string) (field, op string) {
		for _, m := range operatorMap {
			if strings.HasSuffix(part, m.Suffix) {
				return part[:len(part)-len(m.Suffix)], m.SQLOp
			}
		}
		return part, "="
	}

	// Tách các phần hậu tố
	methodName, orderByPart := extractSuffixPart(methodName, "OrderBy")
	orderByPart, limitPart := extractSuffixPart(orderByPart, "Limit")

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

	// Parse WHERE
	whereClauses, _, err := parseWhereConditions(methodName, parseFieldOp)
	if err != nil {
		return nil, err
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

		if field.Type.Kind() == reflect.Func {
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

			// Kiểm tra kiểu trả về của hàm động
			if funcType.NumOut() != 2 {
				return fmt.Errorf("method %s phải trả về 2 giá trị (result, error)", methodName)
			}
			if funcType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
				return fmt.Errorf("method %s output cuối cùng phải là error", methodName)
			}
			if isFindAll {
				if funcType.Out(0).Kind() != reflect.Slice {
					return fmt.Errorf("method %s phải trả về slice cho FindAllBy", methodName)
				}
			} else {
				if funcType.Out(0).Kind() != reflect.Ptr {
					return fmt.Errorf("method %s phải trả về pointer cho FindBy", methodName)
				}
			}

			fn := reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
				ctxVal := args[0]
				ctx := ctxVal.Interface().(context.Context)

				params := make([]interface{}, len(args)-1)
				for i := 1; i < len(args); i++ {
					params[i-1] = args[i].Interface()
				}

				// Đếm số lượng ? trong where clause
				numPlaceholders := strings.Count(strings.Join(qp.WhereClauses, " "), "?")
				if len(params) != numPlaceholders {
					results = make([]reflect.Value, funcType.NumOut())
					err := fmt.Errorf("số lượng tham số truyền vào (%d) không khớp với số lượng điều kiện (%d)", len(params), numPlaceholders)
					for i := 0; i < funcType.NumOut()-1; i++ {
						results[i] = reflect.Zero(funcType.Out(i))
					}
					results[len(results)-1] = reflect.ValueOf(err)
					return results
				}

				dbWithCtx := r.DataSource.WithContext(ctx).Model(new(T))

				if isFindAll {
					var res []T
					err := buildGormQuery(dbWithCtx, qp, params).Find(&res).Error
					if err != nil {
						results = []reflect.Value{
							reflect.MakeSlice(funcType.Out(0), 0, 0),
							reflect.ValueOf(err),
						}
						return results
					}
					results = []reflect.Value{
						reflect.ValueOf(res),
						reflect.Zero(funcType.Out(1)), // nil error
					}
				} else {
					// Khởi tạo một con trỏ tới kiểu của T
					resPtr := reflect.New(funcType.Out(0).Elem())
					err := buildGormQuery(dbWithCtx, qp, params).First(resPtr.Interface()).Error
					if err != nil {
						results = []reflect.Value{
							reflect.Zero(funcType.Out(0)),
							reflect.ValueOf(err),
						}
						return results
					}
					results = []reflect.Value{
						resPtr,
						reflect.Zero(funcType.Out(1)),
					}
				}

				return
			})

			v.Field(i).Set(fn)
		}
	}
	return nil
}
