package pbmoney

import (
	pb "google.golang.org/genproto/googleapis/type/money"
	"reflect"
	"testing"
)

func mmc(u int64, n int32, c string) pb.Money { return pb.Money{Units: u, Nanos: n, CurrencyCode: c} }
func mm(u int64, n int32) pb.Money            { return mmc(u, n, "") }

func TestSum(t *testing.T) {
	type args struct {
		l pb.Money
		r pb.Money
	}
	tests := []struct {
		name    string
		args    args
		want    pb.Money
		wantErr error
	}{
		{"0+0=0", args{mm(0, 0), mm(0, 0)}, mm(0, 0), nil},
		{"Error: currency code on left", args{mmc(0, 0, "XXX"), mm(0, 0)}, mm(0, 0), ErrMismatchingCurrency},
		{"Error: currency code on right", args{mm(0, 0), mmc(0, 0, "YYY")}, mm(0, 0), ErrMismatchingCurrency},
		{"Error: currency code mismatch", args{mmc(0, 0, "AAA"), mmc(0, 0, "BBB")}, mm(0, 0), ErrMismatchingCurrency},
		{"Error: invalid +/-", args{mm(+1, -1), mm(0, 0)}, mm(0, 0), ErrInvalidValue},
		{"Error: invalid -/+", args{mm(0, 0), mm(-1, +2)}, mm(0, 0), ErrInvalidValue},
		{"Error: invalid nanos", args{mm(0, 1000000000), mm(1, 0)}, mm(0, 0), ErrInvalidValue},
		{"both positive (no carry)", args{mm(2, 200000000), mm(2, 200000000)}, mm(4, 400000000), nil},
		{"both positive (nanos=max)", args{mm(2, 111111111), mm(2, 888888888)}, mm(4, 999999999), nil},
		{"both positive (carry)", args{mm(2, 200000000), mm(2, 900000000)}, mm(5, 100000000), nil},
		{"both negative (no carry)", args{mm(-2, -200000000), mm(-2, -200000000)}, mm(-4, -400000000), nil},
		{"both negative (carry)", args{mm(-2, -200000000), mm(-2, -900000000)}, mm(-5, -100000000), nil},
		{"mixed (larger positive, just decimals)", args{mm(11, 0), mm(-2, 0)}, mm(9, 0), nil},
		{"mixed (larger negative, just decimals)", args{mm(-11, 0), mm(2, 0)}, mm(-9, 0), nil},
		{"mixed (larger positive, no borrow)", args{mm(11, 100000000), mm(-2, -100000000)}, mm(9, 0), nil},
		{"mixed (larger positive, with borrow)", args{mm(11, 100000000), mm(-2, -9000000 /*.09*/)}, mm(9, 91000000 /*.091*/), nil},
		{"mixed (larger negative, no borrow)", args{mm(-11, -100000000), mm(2, 100000000)}, mm(-9, 0), nil},
		{"mixed (larger negative, with borrow)", args{mm(-11, -100000000), mm(2, 9000000 /*.09*/)}, mm(-9, -91000000 /*.091*/), nil},
		{"0+negative", args{mm(0, 0), mm(-2, -100000000)}, mm(-2, -100000000), nil},
		{"negative+0", args{mm(-2, -100000000), mm(0, 0)}, mm(-2, -100000000), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sum(tt.args.l, tt.args.r)
			if err != tt.wantErr {
				t.Errorf("Sum([%v],[%v]): expected err=\"%v\" got=\"%v\"", tt.args.l, tt.args.r, tt.wantErr, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sum([%v],[%v]) = %v, want %v", tt.args.l, tt.args.r, got, tt.want)
			}
		})
	}
}
