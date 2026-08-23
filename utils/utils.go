package utils

import (
	"strconv"
	"time"
)

type DriverValue struct {
	Value string
	IsNull bool
}
func Serialize(params []any) []DriverValue {
	sv := make([]DriverValue, len(params))
	for i, v := range params {
		if v == nil {
			sv[i] = DriverValue{IsNull: true}
		}
		switch t := v.(type) {
		case string:
			sv[i] = DriverValue{Value: t, IsNull: false}
		case *string:
			if t == nil {
				sv[i] = DriverValue{IsNull: true}
			} else {
				sv[i] = DriverValue{Value: *t, IsNull: false}
			}
		case int:
			sv[i] = DriverValue{Value: strconv.Itoa(t), IsNull: false}
		case *int:
			if t == nil {
				sv[i] = DriverValue{IsNull: true}
			} else {
				sv[i] = DriverValue{Value: strconv.Itoa(*t), IsNull: false}
			}
		case bool:
			if t {
				sv[i] = DriverValue{Value: "t", IsNull: false}
			} else {
				sv[i] = DriverValue{Value: "f", IsNull: false}
			}
		case *bool:
			if t == nil { sv[i] = DriverValue{IsNull: true} } else { 
				if *t { 
					sv[i] = DriverValue{Value: "t", IsNull: false}
				} else { 
					sv[i] = DriverValue{Value: "f", IsNull: false}
				} 
			}
		case time.Time:
			sv[i] = DriverValue{Value: t.Format("2006-01-02"), IsNull: false}
		case *time.Time:
			if t == nil { sv[i] = DriverValue{IsNull: true} } else { sv[i] = DriverValue{Value: t.Format("2006-01-02"), IsNull: false} }
		}

	}
	return sv
}