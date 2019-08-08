package mymodel

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	// DBTag ...
	DBTag = "db"

	// Combine
	CombineADD = "ADD"
	CombineOR  = "OR"

	// ISO8601Date ISO 8601 format with just the date
	ISO8601Date = "2006-01-02"

	// SQLDatetime YYYY-MM-DD HH:II:SS format with the date and time
	SQLDatetime = "2006-01-02 15:04:05"

	// TimestampFormat Timestamp Format
	TimestampFormat = "20060102150405"

	// MaxLimit ...
	MaxLimit = 50

	// MinOffset ...
	MinOffset = 0

	// SQLNullDate default null date
	SQLNullDate = "0000-00-00 00:00:00"

	// SQLNoDatabaseConnection no database connection
	SQLNoDatabaseConnection = "no database connection"

	// NoInsertRecordProvided no insert record provided
	NoInsertRecordProvided = "no insert record provided"

	// NoUpdateRecordProvided no insert record provided
	NoUpdateRecordProvided = "no update record provided"

	// SQLInvalidOperator invalid sql operator
	SQLInvalidOperator = "invalid sql operator"

	//LimitRangeErrorCode limit must be a value 1 to 50
	LimitRangeErrorCode = "limit must be a value 1 to 50"

	//OffsetRangeErrorCode offset must be >= 0
	OffsetRangeErrorCode = "offset must be >= 0"

	// SQLNoRowsErrorCode sql: no rows in result set
	SQLNoRowsErrorCode = "sql: no rows in result set"

	InvalidINClauseValue = "IN clause values should be a slice of interface"

	InvalidBetweenClauseValue = "value should be a slice of interface in case of between clause"

	// DateField ...
	DateField = "date"

	// JSONTag ...
	JSONTag = "json"

	// TypeString string
	TypeString = "string"

	// TypeInt int
	TypeInt = "int"

	// TypeFloat64 float64
	TypeFloat64 = "float64"
)

// Operators
const (
	OperatorEqual            = "="
	OperatorNoEqual          = "!="
	OperatorGreaterThan      = ">"
	OperatorGreaterThanEqual = ">="
	OperatorLessThan         = "<"
	OperatorLeasThanEqual    = "<="
	OperatorIN               = "IN"
	OperatorIsNull           = "IS NULL"
	OperatorIsNotNull        = "IS NOT NULL"
	OperatorLike             = "LIKE"
	OperatorBetween          = "BETWEEN"
)

// Model represents the core model
type Model struct {
	DB        *sqlx.DB `db:"-" json:"-"`
	Limit     int      `db:"-" json:"-"`
	Offset    int      `db:"-" json:"-"`
	SortOrder []string `db:"-" json:"-"`
	CacheThis bool     `db:"-" json:"-"`
	TableName string   `db:"-" json:"-"`
}

// Condition ...
type Condition struct {
	Combine  string
	Field    string
	Operator string
	Value    interface{} // In case of BETWEEN it will be slice
}

// Conditions ...
type Conditions []Condition

// SortDirection ...
var SortDirection = map[string]string{
	"-": "ASC",
	"+": "DESC",
}

// UnixTimestamp return utc timestamp
func UnixTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Local().Unix())
}

// UnixToMysqlTime return utc timestamp
func UnixToMysqlTime(sec string, nsec string) string {
	iSec, _ := strconv.ParseInt(sec, 10, 64)
	iNsec, _ := strconv.ParseInt(nsec, 10, 64)

	return time.Unix(iSec, iNsec).Format(SQLDatetime)
}

// Select ...
func (m *Model) Select(dest interface{}, conditions Conditions) (err error) {
	var (
		sql, whereClause string
		args             []interface{}
		columns, order   []string
	)

	columns = m.getColumns(dest)
	sql = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), m.TableName)

	if whereClause, args, err = m.getWhereClause(conditions); err != nil {
		return
	}

	if whereClause != "" {
		sql += whereClause
	}

	if len(m.SortOrder) > 0 {

		for _, s := range m.SortOrder {
			d, c := string(s[0]), string(s[1:])
			order = append(order, c+" "+SortDirection[d])

		}
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(order, ","))
	}

	if m.Limit != 0 {
		sql += " LIMIT " + strconv.Itoa(m.Offset) + "," + strconv.Itoa(m.Limit)
	}

	return m.DB.Select(dest, sql, args...)
}

// SelectComplex ...
func (m *Model) SelectComplex(dest interface{}) {

}

// Insert - To insert single or multiple records
func (m *Model) Insert(insertSet []map[string]interface{}) (res sql.Result, err error) {
	var (
		rows, args   []interface{}
		tRecCount    int
		columns, val []string
		sql, query   string
	)

	tRecCount = len(insertSet)

	if tRecCount == 0 {
		err = errors.New(NoInsertRecordProvided)
		return
	}

	// prepare values place holders
	for _, r := range insertSet {
		rows = append(rows, r)
		val = append(val, "(?)")
	}

	// fetch columns
	for c, _ := range insertSet[0] {
		columns = append(columns, c)
	}

	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", m.TableName, strings.Join(columns, ","), strings.Join(val, ","))

	query, args, err = sqlx.In(sql, rows...)
	if err != nil {
		return
	}

	res, err = m.DB.Exec(query, args)

	return
}

// Update -
func (m *Model) Update(set map[string]interface{}, conditions map[string]interface{}) (res sql.Result, err error) {

	var (
		args                    []interface{}
		updateSet, conditionSet []string
		query                   string
	)

	if set != nil {
		err = errors.New(NoUpdateRecordProvided)
		return
	}

	for c, v := range set {
		updateSet = append(updateSet, c+" = ?")
		args = append(args, v)
	}

	query = fmt.Sprintf("UPDATE %s SET %s", m.TableName, strings.Join(updateSet, ","))

	if conditions != nil {
		for c, v := range conditions {
			conditionSet = append(conditionSet, c+" = ?")
			args = append(args, v)
		}
		if len(conditionSet) > 0 {
			query += " WHERE " + strings.Join(conditionSet, ",")
		}
	}

	res, err = m.DB.Exec(query, args...)
	return
}

// Delete ...
func (m *Model) Delete(query string, args ...interface{}) (res sql.Result, err error) {
	//TODO:
	res, err = m.DB.Exec(query, args...)
	return
}

// getWhereClause ...
func (m *Model) getWhereClause(conditions Conditions) (cond string, args []interface{}, err error) {

	var (
		conds []string
	)
	if len(conditions) == 0 {
		return
	}

	for _, c := range conditions {
		switch c.Operator {
		case OperatorEqual, OperatorNoEqual, OperatorGreaterThan, OperatorGreaterThanEqual, OperatorLessThan, OperatorLeasThanEqual:
			conds = append(conds, c.Combine+" "+c.Field+""+c.Operator+"?")
			args = append(args, c.Value)
			break
		case OperatorIN:
			v := reflect.ValueOf(c.Value).Type().Kind().String()
			if v != "slice" {
				err = errors.New(InvalidINClauseValue)
				return
			}
			bindVars := strings.Trim(strings.Repeat(`?,`, len(c.Value.([]interface{}))), ",")

			t := fmt.Sprintf(c.Combine+` %s IN (%s)`, c.Field, bindVars)
			conds = append(conds, t)
			args = append(args, c.Value.([]interface{})...)
			break
		case OperatorIsNull, OperatorIsNotNull:
			conds = append(conds, c.Combine+` `+c.Field+` `+c.Operator)
			break
		case OperatorLike:
			conds = append(conds, c.Combine+` `+c.Field+` `+c.Operator+` "`+c.Value.(string)+`"`) // Example: Go% , %ang
			break
		case OperatorBetween:
			v := reflect.ValueOf(c.Value).Type().Kind().String()
			if v != "slice" {
				err = errors.New(InvalidBetweenClauseValue)
				return
			}

			vals := c.Value.([]interface{})
			clause := fmt.Sprintf(c.Combine+" %s BETWEEN ? AND ?", c.Field)
			conds = append(conds, clause)
			args = append(args, vals...)
			break
		default:
			break

		}
	}

	if len(conds) > 0 {
		cond = fmt.Sprintf(" WHERE %s", strings.Join(conds, " "))
	}

	return
}

// getColumns ...
func (m *Model) getColumns(dest interface{}) (columns []string) {

	v := reflect.ValueOf(dest).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		dbColumn := f.Tag.Get(DBTag)
		if dbColumn != "" {
			columns = append(columns, dbColumn)
		}
	}

	return
}
