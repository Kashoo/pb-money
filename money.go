package pbmoney

import (
	"errors"
	"fmt"
	pb "google.golang.org/genproto/googleapis/type/money"
	"strconv"
	"strings"
)

const (
	nanosMin = -999999999
	nanosMax = +999999999
	nanosMod = 1000000000
)

var (
	// ErrInvalidValue is returned if the specified money amount is not valid.
	ErrInvalidValue = errors.New("one of the specified money values is invalid")

	// ErrMismatchingCurrency is returned if two values don't have the same currency code.
	ErrMismatchingCurrency = errors.New("mismatching currency codes")
)

// IsValid checks if specified value has a valid units/nanos signs and ranges.
func IsValid(m *pb.Money) bool {
	return signMatches(m) && validNanos(m.GetNanos())
}

func signMatches(m *pb.Money) bool {
	return m.GetNanos() == 0 || m.GetUnits() == 0 || (m.GetNanos() < 0) == (m.GetUnits() < 0)
}

func validNanos(nanos int32) bool { return nanosMin <= nanos && nanos <= nanosMax }

// IsZero returns true if the specified money value is equal to zero.
func IsZero(m *pb.Money) bool { return m.GetUnits() == 0 && m.GetNanos() == 0 }

// IsPositive returns true if the specified money value is valid and is
// positive.
func IsPositive(m *pb.Money) bool {
	return IsValid(m) && m.GetUnits() > 0 || (m.GetUnits() == 0 && m.GetNanos() > 0)
}

// IsNegative returns true if the specified money value is valid and is
// negative.
func IsNegative(m *pb.Money) bool {
	return IsValid(m) && m.GetUnits() < 0 || (m.GetUnits() == 0 && m.GetNanos() < 0)
}

// AreSameCurrency returns true if values l and r have a currency code and
// they are the same values.
func AreSameCurrency(l, r *pb.Money) bool {
	return l.GetCurrencyCode() == r.GetCurrencyCode() && l.GetCurrencyCode() != ""
}

// AreEquals returns true if values l and r are the equal, including the
// currency. This does not check validity of the provided values.
func AreEquals(l, r *pb.Money) bool {
	return l.GetCurrencyCode() == r.GetCurrencyCode() &&
		l.GetUnits() == r.GetUnits() && l.GetNanos() == r.GetNanos()
}

// Negate returns the same amount with the sign negated.
func Negate(m *pb.Money) *pb.Money {
	return &pb.Money{
		Units:        -m.GetUnits(),
		Nanos:        -m.GetNanos(),
		CurrencyCode: m.GetCurrencyCode()}
}

// Must panics if the given error is not nil. This can be used with other
// functions like: "m := Must(Sum(a,b))".
func Must(v *pb.Money, err error) *pb.Money {
	if err != nil {
		panic(err)
	}
	return v
}

// Sum adds two values. Returns an error if one of the values are invalid or
// currency codes are not matching (unless currency code is unspecified for
// both).
func Sum(l, r *pb.Money) (*pb.Money, error) {
	if !IsValid(l) || !IsValid(r) {
		return &pb.Money{}, ErrInvalidValue
	} else if l.GetCurrencyCode() != r.GetCurrencyCode() {
		return &pb.Money{}, ErrMismatchingCurrency
	}
	units := l.GetUnits() + r.GetUnits()
	nanos := l.GetNanos() + r.GetNanos()

	if (units == 0 && nanos == 0) || (units > 0 && nanos >= 0) || (units < 0 && nanos <= 0) {
		// same sign <units, nanos>
		units += int64(nanos / nanosMod)
		nanos = nanos % nanosMod
	} else {
		// different sign. nanos guaranteed to not to go over the limit
		if units > 0 {
			units--
			nanos += nanosMod
		} else {
			units++
			nanos -= nanosMod
		}
	}

	return &pb.Money{
		Units:        units,
		Nanos:        nanos,
		CurrencyCode: l.GetCurrencyCode()}, nil
}

// MultiplyInt is a slow multiplication operation done through adding the value
// to itself n-1 times.
func MultiplyInt(m *pb.Money, n uint32) *pb.Money {
	out := m
	for n > 1 {
		out = Must(Sum(out, m))
		n--
	}
	return out
}

// DivideInt is a slow multiplication operation done through adding the value
// to itself n-1 times.
func DivideInt(m *pb.Money, n uint32) *pb.Money {
	out := Negate(m)
	for n > 1 {
		out = Must(Sum(out, m))
		n--
	}
	return out
}

func MultipleFast(l, r *pb.Money) *pb.Money {
	lr := unitsAndNanoPartToMicros(l.Units, l.Nanos)
	rr := unitsAndNanoPartToMicros(r.Units, r.Nanos)
	ln := lr * rr
	return toGoogleMoney(ln, l.CurrencyCode)
}

func DivideFast(l, r *pb.Money) *pb.Money {
	lr := unitsAndNanoPartToMicros(l.Units, l.Nanos)
	rr := unitsAndNanoPartToMicros(r.Units, r.Nanos)

	ln := lr / rr
	return toGoogleMoney(ln, l.CurrencyCode)
}

func DivideFastInt(l *pb.Money, r int64) *pb.Money {
	lr := unitsAndNanoPartToMicros(l.Units, l.Nanos)

	ln := lr / r
	return toGoogleMoney(ln, l.CurrencyCode)
}
func MultipleFastInt(l *pb.Money, r int64) *pb.Money {
	lr := unitsAndNanoPartToMicros(l.Units, l.Nanos)
	ln := lr * r
	return toGoogleMoney(ln, l.CurrencyCode)
}

func toGoogleMoney(valueMicros int64, currencyCode string) *pb.Money {
	units, nanoPart := microsToUnitsAndNanoPart(valueMicros)
	return &pb.Money{
		CurrencyCode: currencyCode,
		Units:        units,
		Nanos:        nanoPart,
	}
}

func unitsAndMicroPartToMicros(units int64, micros int64) int64 {
	return unitsToMicros(units) + micros
}

//func unitsAndNanoPartToMicros(units int64, nanos int32) int64 {
//	return unitsToMicros(units) + int64(nanos/1000)
//}

func microsToUnitsAndMicroPart(micros int64) (int64, int64) {
	return micros / 1000000, micros % 1000000
}

// edited
func unitsAndNanoPartToMicros(units int64, nanos int32) int64 {
	return (units*100 + int64(nanos/10000000))
}

//func microsToUnitsAndNanoPart(micros int64) (int64, int32) {
//	return micros / 1000000, int32(micros%1000000) * 1000
//}
// edited
func microsToUnitsAndNanoPart(micros int64) (int64, int32) {
	return micros / 10000, int32(micros%10000) * 100000
}

func unitsToMicros(units int64) int64 {
	return units * 1000000
}

func floatUnitsToMicros(floatUnits float64) int64 {
	return int64(floatUnits * 1000000.0)
}

func microsToFloat(micros int64) float64 {
	return float64(micros) / 1000000.0
}

func ToStringDollars(l *pb.Money) string {

	nanos := strconv.Itoa(int(l.GetNanos()))
	nanos = strings.TrimRight(nanos, "0")
	if nanos == "" {
		nanos = "00"
	} else if len(nanos) == 1 {
		nanos = nanos + "0"
	}
	return fmt.Sprintf("%d.%d", l.GetUnits(), nanos)
}

func ToInt(l *pb.Money) int64 {
	return unitsAndNanoPartToMicros(l.GetUnits(), l.GetNanos())
}
