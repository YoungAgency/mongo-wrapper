package codec

import (
	"fmt"
	"math"
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DecimalCodec is used to encode decimal.Decimal types in primitive.Decimal128 and
// decode primitive.Decimal128 in decimal.Decimal
// if WriteInt is true, then it will write int32 or int64 for integer values
// if AllowWriteInt32 is true, then it will write int32 for integer values if possible
type DecimalCodec struct {
	WriteInt        bool
	AllowWriteInt32 bool
}

// EncodeValue implements bsoncodec.ValueEncoder interface
func (dc DecimalCodec) EncodeValue(ctx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bsoncodec.ValueEncoderError{
			Name:     "DecimalEncodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}

	dec := val.Interface().(decimal.Decimal)
	if dc.WriteInt && dec.IsInteger() {
		i64 := dec.IntPart()
		if dc.AllowWriteInt32 && i64 > math.MinInt32 && i64 < math.MaxInt32 {
			return vw.WriteInt32(int32(i64))
		} else {
			return vw.WriteInt64(i64)
		}
	} else {
		mongoDecimal, err := decimalToPrimitive128(dec)
		if err != nil {
			return fmt.Errorf("DecimalCodec: unable to convert decimal.Decimal to primitive.Decimal128, %v", err)
		}
		return vw.WriteDecimal128(mongoDecimal)
	}
}

// DecodeValue implements bsoncodec.ValueEncoder interface
func (dc DecimalCodec) DecodeValue(ctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != reflect.TypeOf(decimal.Decimal{}) {
		return bsoncodec.ValueDecoderError{
			Name:     "DecimalDecodeValue",
			Types:    []reflect.Type{reflect.TypeOf(decimal.Decimal{})},
			Received: val,
		}
	}
	switch vr.Type() {
	case bsontype.Decimal128:
		mongoDecimal, err := vr.ReadDecimal128()
		if err != nil {
			return err
		}
		newDec, err := primitive128ToDecimal(mongoDecimal)
		if err != nil {
			return fmt.Errorf("DecimalCodec: unable to convert primitive.Decimal128 to decimal.Decimal %v", err)
		}
		val.Set(reflect.ValueOf(newDec))
	case bsontype.Int32:
		i32, err := vr.ReadInt32()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(decimal.NewFromInt32(i32)))
	case bsontype.Int64:
		i64, err := vr.ReadInt64()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(decimal.NewFromInt(i64)))
	case bsontype.Double:
		f64, err := vr.ReadDouble()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(decimal.NewFromFloat(f64)))
	case bsontype.Null:
		val.Set(reflect.ValueOf(decimal.Zero))
	case bsontype.String:
		str, err := vr.ReadString()
		if err != nil {
			return err
		}
		newDec, err := decimal.NewFromString(str)
		if err != nil {
			return fmt.Errorf("DecimalCodec: unable to convert string to decimal.Decimal %v", err)
		}
		val.Set(reflect.ValueOf(newDec))
	default:
		return fmt.Errorf("cannot decode %v into a decimal.Decimal type", vr.Type())
	}
	return nil
}

func decimalToPrimitive128(d decimal.Decimal) (primitive.Decimal128, error) {
	coefficient, exp := d.Coefficient(), d.Exponent()
	mongoDecimal, ok := primitive.ParseDecimal128FromBigInt(coefficient, int(exp))
	if !ok {
		return primitive.Decimal128{}, fmt.Errorf("unable to parse Decimal128 from big int")
	}
	return mongoDecimal, nil
}

func primitive128ToDecimal(p primitive.Decimal128) (decimal.Decimal, error) {
	bigInt, exp, err := p.BigInt()
	if err != nil {
		return decimal.Decimal{}, err
	}
	// convert exp to int32 should never be a problem since exp is in [-6176, 6111]
	d := decimal.NewFromBigInt(bigInt, int32(exp))
	return d, nil
}
