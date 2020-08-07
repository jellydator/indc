package indc

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_NewAroon(t *testing.T) {
	cc := map[string]struct {
		Trend  Trend
		Length int
		Offset int
		Result Aroon
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Trend:  TrendDown,
			Length: 5,
			Offset: 2,
			Result: Aroon{trend: TrendDown, length: 5, offset: 2, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewAroon(c.Trend, c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_Aroon_Equal(t *testing.T) {
	assert.True(t, Aroon{trend: TrendUp, length: 3, offset: 2}.equal(Aroon{trend: TrendUp, length: 3, offset: 2}))
	assert.False(t, Aroon{trend: TrendDown, length: 3, offset: 2}.equal(Aroon{trend: TrendUp, length: 3, offset: 2}))
	assert.False(t, Aroon{trend: TrendDown, length: 3, offset: 2}.equal(SMA{}))
}

func Test_Aroon_Trend(t *testing.T) {
	assert.Equal(t, TrendUp, Aroon{trend: TrendUp}.Trend())
}

func Test_Aroon_Length(t *testing.T) {
	assert.Equal(t, 1, Aroon{length: 1}.Length())
}

func Test_Aroon_Offset(t *testing.T) {
	assert.Equal(t, 3, Aroon{offset: 3}.Offset())
}

func Test_Aroon_validate(t *testing.T) {
	cc := map[string]struct {
		Aroon Aroon
		Error error
		Valid bool
	}{
		"Invalid trend": {
			Aroon: Aroon{trend: 70, length: 5, offset: 0},
			Error: ErrInvalidTrend,
			Valid: false,
		},
		"Invalid length": {
			Aroon: Aroon{trend: TrendDown, length: 0, offset: 0},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			Aroon: Aroon{trend: TrendDown, length: 1, offset: -1},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			Aroon: Aroon{trend: TrendUp, length: 1, offset: 0},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.Aroon.validate())
			assert.Equal(t, c.Valid, c.Aroon.valid)
		})
	}
}

func Test_Aroon_Calc(t *testing.T) {
	cc := map[string]struct {
		Aroon  Aroon
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			Aroon: Aroon{valid: false},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			Aroon: Aroon{trend: TrendDown, length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with TrendUp": {
			Aroon: Aroon{trend: TrendUp, length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(25),
				decimal.NewFromInt(31),
				decimal.NewFromInt(38),
				decimal.NewFromInt(35),
				decimal.NewFromInt(29),
				decimal.NewFromInt(29),
			},
			Result: decimal.NewFromInt(40),
		},
		"Successful calculation with TrendDown": {
			Aroon: Aroon{trend: TrendDown, length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(25),
				decimal.NewFromInt(31),
				decimal.NewFromInt(38),
				decimal.NewFromInt(35),
				decimal.NewFromInt(29),
				decimal.NewFromInt(29),
			},
			Result: Hundred,
		},
		"Successful calculation with offset": {
			Aroon: Aroon{trend: TrendDown, length: 5, offset: 3, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(25),
				decimal.NewFromInt(31),
				decimal.NewFromInt(38),
				decimal.NewFromInt(35),
				decimal.NewFromInt(29),
				decimal.NewFromInt(29),
				decimal.NewFromInt(35),
				decimal.NewFromInt(29),
				decimal.NewFromInt(29),
			},
			Result: Hundred,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.Aroon.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_Aroon_Count(t *testing.T) {
	assert.Equal(t, 8, Aroon{length: 5, offset: 3}.Count())
}

func Test_Aroon_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result Aroon
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"trend":"upp","length":1,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"trend":"up","length":1,"offset":0}`,
			Result: Aroon{trend: TrendUp, length: 1, offset: 0, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := Aroon{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_Aroon_MarshalJSON(t *testing.T) {
	d, err := Aroon{trend: TrendDown, length: 1, offset: 4}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"trend":"down","length":1,"offset":4}`, string(d))
}

func Test_Aroon_namedMarshalJSON(t *testing.T) {
	d, err := Aroon{trend: TrendDown, length: 1, offset: 4}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"aroon","trend":"down","length":1,"offset":4}`, string(d))
}

func Test_NewBB(t *testing.T) {
	cc := map[string]struct {
		Band    Band
		StdDevs decimal.Decimal
		Length  int
		Offset  int
		Result  BB
		Error   error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Band:    BandUpper,
			StdDevs: decimal.RequireFromString("2.5"),
			Length:  5,
			Offset:  2,
			Result:  BB{band: BandUpper, stdDevs: decimal.RequireFromString("2.5"), length: 5, offset: 2, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewBB(c.Band, c.StdDevs, c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_BB_Equal(t *testing.T) {
	assert.True(t, BB{band: BandUpper, stdDevs: decimal.RequireFromString("2"), length: 3, offset: 2}.equal(BB{band: BandUpper, stdDevs: decimal.RequireFromString("2"), length: 3, offset: 2}))
	assert.False(t, BB{band: BandUpper, stdDevs: decimal.RequireFromString("2"), length: 3, offset: 2}.equal(BB{band: BandUpper, stdDevs: decimal.RequireFromString("3"), length: 3, offset: 2}))
	assert.False(t, BB{band: BandUpper, length: 3, offset: 2}.equal(SMA{}))
}

func Test_BB_Band(t *testing.T) {
	assert.Equal(t, BandLower, BB{band: BandLower}.Band())
}

func Test_BB_StandardDeviations(t *testing.T) {
	assert.Equal(t, decimal.RequireFromString("2"), BB{stdDevs: decimal.RequireFromString("2")}.StandardDeviations())
}

func Test_BB_Length(t *testing.T) {
	assert.Equal(t, 1, BB{length: 1}.Length())
}

func Test_BB_Offset(t *testing.T) {
	assert.Equal(t, 3, BB{offset: 3}.Offset())
}

func Test_BB_validate(t *testing.T) {
	cc := map[string]struct {
		BB    BB
		Error error
		Valid bool
	}{
		"Invalid band": {
			BB:    BB{band: 70, length: 5, offset: 0},
			Error: ErrInvalidBand,
			Valid: false,
		},
		"Invalid length": {
			BB:    BB{band: BandUpper, length: 0, offset: 0},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			BB:    BB{band: BandUpper, length: 1, offset: -1},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			BB:    BB{band: BandUpper, length: 1, offset: 0},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.BB.validate())
			assert.Equal(t, c.Valid, c.BB.valid)
		})
	}
}

func Test_BB_Calc(t *testing.T) {
	cc := map[string]struct {
		BB     BB
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			BB:    BB{valid: false},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			BB: BB{band: BandUpper, length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with BandUpper": {
			BB: BB{band: BandUpper, length: 5, stdDevs: decimal.RequireFromString("1"), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35).Add(sqrt(decimal.RequireFromString("13.6"))),
		},
		"Successful calculation with BandUpper using offset": {
			BB: BB{band: BandUpper, length: 5, stdDevs: decimal.RequireFromString("1"), offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35).Add(sqrt(decimal.RequireFromString("13.6"))),
		},
		"Successful calculation with BandMiddle": {
			BB: BB{band: BandMiddle, length: 5, stdDevs: decimal.RequireFromString("1"), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35),
		},
		"Successful calculation with BandMiddle using offset": {
			BB: BB{band: BandMiddle, length: 5, stdDevs: decimal.RequireFromString("1"), offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35),
		},
		"Successful calculation with BandLower": {
			BB: BB{band: BandLower, length: 5, stdDevs: decimal.RequireFromString("2.5"), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35).Sub(sqrt(decimal.RequireFromString("13.6")).Mul(decimal.RequireFromString("2.5"))),
		},
		"Successful calculation with BandLower using offset": {
			BB: BB{band: BandLower, length: 5, stdDevs: decimal.RequireFromString("2"), offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(35),
				decimal.NewFromInt(40),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
				decimal.NewFromInt(38),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(35).Sub(sqrt(decimal.RequireFromString("13.6")).Mul(decimal.RequireFromString("2"))),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.BB.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_BB_Count(t *testing.T) {
	assert.Equal(t, 2, BB{length: 1, offset: 1}.Count())
}

func Test_BB_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result BB
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"band":"brand","length":1,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"band":"lower","standard_deviations":"3","length":2,"offset":4}`,
			Result: BB{band: BandLower, stdDevs: decimal.RequireFromString("3"), length: 2, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := BB{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_BB_MarshalJSON(t *testing.T) {
	d, err := BB{band: BandLower, stdDevs: decimal.RequireFromString("1"), length: 3, offset: 0}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"band":"lower","standard_deviations":"1","length":3,"offset":0}`, string(d))
}

func Test_BB_namedMarshalJSON(t *testing.T) {
	d, err := BB{band: BandUpper, stdDevs: decimal.RequireFromString("2.3"), length: 1, offset: 4}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"bb","band":"upper","standard_deviations":"2.3","length":1,"offset":4}`, string(d))
}

func Test_NewCCI(t *testing.T) {
	cc := map[string]struct {
		Source Indicator
		Factor decimal.Decimal
		Result CCI
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation (default factor)": {
			Source: &IndicatorMock{},
			Factor: decimal.Zero,
			Result: CCI{source: &IndicatorMock{}, factor: decimal.RequireFromString("0.015"), valid: true},
		},
		"Successful creation": {
			Source: &IndicatorMock{},
			Factor: Hundred,
			Result: CCI{source: &IndicatorMock{}, factor: Hundred, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewCCI(c.Source, c.Factor)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_CCI_Equal(t *testing.T) {
	assert.True(t, CCI{source: SMA{}, factor: decimal.Zero}.equal(CCI{source: SMA{}, factor: decimal.Zero}))
	assert.False(t, CCI{source: SMA{}, factor: decimal.NewFromInt(1)}.equal(CCI{source: SMA{}, factor: decimal.Zero}))
	assert.False(t, CCI{source: SMA{length: 1}, factor: decimal.Zero}.equal(CCI{source: SMA{}, factor: decimal.Zero}))
	assert.False(t, CCI{source: SMA{length: 1}, factor: decimal.Zero}.equal(SMA{}))
}

func Test_CCI_Sub(t *testing.T) {
	assert.Equal(t, &IndicatorMock{}, CCI{source: &IndicatorMock{}}.Sub())
}

func Test_CCI_Factor(t *testing.T) {
	assert.Equal(t, Hundred, CCI{factor: Hundred}.Factor())
}

func Test_CCI_Offset(t *testing.T) {
	assert.Equal(t, 4, CCI{source: SMA{offset: 4}}.Offset())
}

func Test_CCI_validate(t *testing.T) {
	cc := map[string]struct {
		CCI   CCI
		Error error
		Valid bool
	}{
		"Invalid source": {
			CCI:   CCI{source: nil},
			Error: ErrInvalidSource,
			Valid: false,
		},
		"Invalid factor": {
			CCI:   CCI{source: &IndicatorMock{}, factor: decimal.NewFromInt(-1)},
			Error: errors.New("invalid factor"),
			Valid: false,
		},
		"Successful validation": {
			CCI:   CCI{source: &IndicatorMock{}, factor: decimal.RequireFromString("1")},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.CCI.validate())
			assert.Equal(t, c.Valid, c.CCI.valid)
		})
	}
}

func Test_CCI_Calc(t *testing.T) {
	stubIndicator := func(v decimal.Decimal, e error, a int) *IndicatorMock {
		return &IndicatorMock{
			CalcFunc: func(dd []decimal.Decimal) (decimal.Decimal, error) {
				return v, e
			},
			CountFunc: func() int {
				return a
			},
		}
	}

	cc := map[string]struct {
		CCI    CCI
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			CCI:   CCI{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			CCI: CCI{source: stubIndicator(decimal.Zero, nil, 10), factor: decimal.RequireFromString("0.015"), valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Invalid source calc": {
			CCI: CCI{source: stubIndicator(decimal.Zero, assert.AnError, 1), factor: decimal.RequireFromString("0.015"), valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: assert.AnError,
		},
		"Successful handled division by 0": {
			CCI: CCI{source: stubIndicator(decimal.NewFromInt(3), nil, 1), factor: decimal.RequireFromString("0.015"), valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(3),
				decimal.NewFromInt(6),
				decimal.NewFromInt(9),
			},
			Result: decimal.Zero,
		},
		"Successful calculation": {
			CCI: CCI{source: stubIndicator(decimal.NewFromInt(3), nil, 3), factor: decimal.RequireFromString("0.015"), valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(3),
				decimal.NewFromInt(6),
				decimal.NewFromInt(9),
			},
			Result: decimal.NewFromInt(200),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.CCI.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_CCI_Count(t *testing.T) {
	indicator := &IndicatorMock{
		CountFunc: func() int {
			return 10
		},
	}

	assert.Equal(t, 10, CCI{source: indicator}.Count())
}

func Test_CCI_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result CCI
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\-_-/}`,
			Error: assert.AnError,
		},
		"Invalid source": {
			JSON:  `{}`,
			Error: assert.AnError,
		},
		"Invalid factor": {
			JSON:  `{"source":{"name":"sma","length":1,"offset":4},"factor":"abc"}`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"source":{"name":"sma","length":1,"offset":4},"factor":"-2"}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"source":{"name":"sma","length":1,"offset":2},"factor":"1"}`,
			Result: CCI{source: SMA{length: 1, valid: true, offset: 2}, factor: decimal.RequireFromString("1"), valid: true},
		},
		"Successful unmarshal with zero factor": {
			JSON:   `{"source":{"name":"sma","length":1,"offset":4},"factor":"0"}`,
			Result: CCI{source: SMA{length: 1, valid: true, offset: 4}, factor: decimal.RequireFromString("0.015"), valid: true},
		},
		"Successful unmarshal with no factor": {
			JSON:   `{"source":{"name":"sma","length":1,"offset":1}}`,
			Result: CCI{source: SMA{length: 1, valid: true, offset: 1}, factor: decimal.RequireFromString("0.015"), valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := CCI{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_CCI_MarshalJSON(t *testing.T) {
	stubIndicator := func(d []byte, e error) *IndicatorMock {
		return &IndicatorMock{
			namedMarshalJSONFunc: func() ([]byte, error) {
				return d, e
			},
		}
	}

	cc := map[string]struct {
		CCI    CCI
		Result string
		Error  error
	}{
		"Invalid source marshal": {
			CCI:   CCI{source: stubIndicator(nil, assert.AnError)},
			Error: assert.AnError,
		},
		"Successful marshal": {
			CCI:    CCI{source: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), factor: Hundred},
			Result: `{"source":{"name":"indicatormock"},"factor":"100"}`,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			d, err := c.CCI.MarshalJSON()
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.JSONEq(t, c.Result, string(d))
		})
	}
}

func Test_CCI_namedMarshalJSON(t *testing.T) {
	stubIndicator := func(d []byte, e error) *IndicatorMock {
		return &IndicatorMock{
			namedMarshalJSONFunc: func() ([]byte, error) {
				return d, e
			},
		}
	}

	cc := map[string]struct {
		CCI    CCI
		Result string
		Error  error
	}{
		"Error returned during source marshalling": {
			CCI:   CCI{source: stubIndicator(nil, assert.AnError)},
			Error: assert.AnError,
		},
		"Successful marshal": {
			CCI:    CCI{source: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), factor: Hundred},
			Result: `{"name":"cci","source":{"name":"indicatormock"},"factor":"100"}`,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			d, err := c.CCI.namedMarshalJSON()
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.JSONEq(t, c.Result, string(d))
		})
	}
}

func Test_NewDEMA(t *testing.T) {
	cc := map[string]struct {
		EMA    EMA
		Result DEMA
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			EMA:    EMA{SMA{length: 1, valid: true, offset: 4}},
			Result: DEMA{ema: EMA{SMA{length: 1, offset: 4, valid: true}}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewDEMA(c.EMA)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_DEMA_Equal(t *testing.T) {
	assert.True(t, DEMA{ema: EMA{}}.equal(DEMA{ema: EMA{}}))
	assert.False(t, DEMA{ema: EMA{}, valid: true}.equal(DEMA{ema: EMA{}}))
	assert.False(t, DEMA{ema: EMA{SMA{length: 4}}}.equal(DEMA{ema: EMA{}}))
	assert.False(t, DEMA{ema: EMA{}}.equal(SMA{}))
}

func Test_DEMA_EMA(t *testing.T) {
	assert.Equal(t, EMA{SMA{length: 1, offset: 2}}, DEMA{ema: EMA{SMA{length: 1, offset: 2}}}.EMA())
}

func Test_DEMA_Offset(t *testing.T) {
	assert.Equal(t, 2, DEMA{ema: EMA{SMA{offset: 2}}}.Offset())
}

func Test_DEMA_validate(t *testing.T) {
	cc := map[string]struct {
		DEMA  DEMA
		Error error
		Valid bool
	}{
		"Invalid EMA": {
			DEMA:  DEMA{ema: EMA{SMA{length: -1, offset: 2}}},
			Error: assert.AnError,
			Valid: false,
		},
		"Successful validation": {
			DEMA:  DEMA{ema: EMA{SMA{length: 1, offset: 2}}},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.DEMA.validate())
			assert.Equal(t, c.Valid, c.DEMA.valid)
		})
	}
}

func Test_DEMA_Calc(t *testing.T) {
	cc := map[string]struct {
		DEMA   DEMA
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			DEMA:  DEMA{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			DEMA: DEMA{ema: EMA{SMA{length: 3, valid: true, offset: 2}}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with offset": {
			DEMA: DEMA{ema: EMA{SMA{length: 3, valid: true, offset: 2}}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(1),
				decimal.NewFromInt(1),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
			},
			Result: decimal.RequireFromString("6.75"),
		},
		"Successful calculation": {
			DEMA: DEMA{ema: EMA{SMA{length: 3, valid: true, offset: 0}}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(1),
				decimal.NewFromInt(1),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
			},
			Result: decimal.RequireFromString("6.75"),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.DEMA.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_DEMA_Count(t *testing.T) {
	assert.Equal(t, 33, DEMA{ema: EMA{SMA{length: 15, offset: 4}}}.Count())
}

func Test_DEMA_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result DEMA
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"ema":{"length":1,"offset":2}}`,
			Result: DEMA{ema: EMA{SMA{length: 1, offset: 2, valid: true}}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := DEMA{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_DEMA_MarshalJSON(t *testing.T) {
	d, err := DEMA{ema: EMA{SMA{length: 1, offset: 3}}}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"ema":{"length":1,"offset":3}}`, string(d))
}

func Test_DEMA_namedMarshalJSON(t *testing.T) {
	d, err := DEMA{ema: EMA{SMA{length: 1, offset: 3}}}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"dema","ema":{"length":1,"offset":3}}`, string(d))
}

func Test_NewEMA(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result EMA
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 4,
			Result: EMA{SMA{length: 1, offset: 4, valid: true}},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewEMA(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_EMA_Equal(t *testing.T) {
	assert.True(t, EMA{SMA{}}.equal(EMA{SMA{}}))
	assert.False(t, EMA{SMA{length: 2}}.equal(EMA{SMA{}}))
	assert.False(t, EMA{SMA{}}.equal(SMA{}))
}

func Test_EMA_Length(t *testing.T) {
	assert.Equal(t, 1, EMA{SMA{length: 1}}.Length())
}

func Test_EMA_Offset(t *testing.T) {
	assert.Equal(t, 1, EMA{SMA{offset: 1}}.Offset())
}

func Test_EMA_validate(t *testing.T) {
	cc := map[string]struct {
		EMA   EMA
		Error error
		Valid bool
	}{
		"Invalid SMA": {
			EMA:   EMA{SMA{length: -1, offset: 2}},
			Error: assert.AnError,
			Valid: false,
		},
		"Invalid offset": {
			EMA:   EMA{SMA{length: 1, offset: -2}},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			EMA:   EMA{SMA{length: 1, offset: 2}},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.EMA.validate())
			assert.Equal(t, c.Valid, c.EMA.valid)
		})
	}
}

func Test_EMA_Calc(t *testing.T) {
	cc := map[string]struct {
		EMA    EMA
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			EMA:   EMA{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			EMA: EMA{SMA{length: 3, offset: 2, valid: true}},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with offset": {
			EMA: EMA{SMA{length: 3, offset: 2, valid: true}},
			Data: []decimal.Decimal{
				decimal.NewFromInt(31),
				decimal.NewFromInt(1),
				decimal.NewFromInt(1),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
			},
			Result: decimal.RequireFromString("4.75"),
		},
		"Successful calculation": {
			EMA: EMA{SMA{length: 3, offset: 0, valid: true}},
			Data: []decimal.Decimal{
				decimal.NewFromInt(31),
				decimal.NewFromInt(1),
				decimal.NewFromInt(1),
				decimal.NewFromInt(2),
				decimal.NewFromInt(3),
			},
			Result: decimal.RequireFromString("4.75"),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.EMA.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_EMA_CalcNext(t *testing.T) {
	cc := map[string]struct {
		EMA    EMA
		Last   decimal.Decimal
		Next   decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			EMA:   EMA{},
			Error: ErrInvalidIndicator,
		},
		"Successful calculation with offset": {
			EMA:    EMA{SMA{length: 3, offset: 1, valid: true}},
			Last:   decimal.NewFromInt(5),
			Next:   decimal.NewFromInt(5),
			Result: decimal.NewFromInt(5),
		},
		"Successful calculation": {
			EMA:    EMA{SMA{length: 3, offset: 0, valid: true}},
			Last:   decimal.NewFromInt(5),
			Next:   decimal.NewFromInt(5),
			Result: decimal.NewFromInt(5),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.EMA.CalcNext(c.Last, c.Next)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_EMA_Count(t *testing.T) {
	assert.Equal(t, 33, EMA{SMA{length: 15, offset: 4}}.Count())
}

func Test_EMA_multiplier(t *testing.T) {
	assert.Equal(t, decimal.RequireFromString("0.5").String(), EMA{SMA{length: 3}}.multiplier().String())
}

func Test_EMA_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result EMA
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":4}`,
			Result: EMA{SMA{length: 1, offset: 4, valid: true}},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := EMA{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_EMA_MarshalJSON(t *testing.T) {
	d, err := EMA{SMA{length: 1, offset: 2}}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":2}`, string(d))
}

func Test_EMA_namedMarshalJSON(t *testing.T) {
	d, err := EMA{SMA{length: 1, offset: 2}}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"ema","length":1,"offset":2}`, string(d))
}

func Test_NewHMA(t *testing.T) {
	cc := map[string]struct {
		WMA    WMA
		Result HMA
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			WMA:    WMA{length: 2, offset: 2, valid: true},
			Result: HMA{wma: WMA{length: 2, offset: 2, valid: true}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewHMA(c.WMA)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_HMA_Equal(t *testing.T) {
	assert.True(t, HMA{wma: WMA{}}.equal(HMA{wma: WMA{}}))
	assert.False(t, HMA{wma: WMA{}, valid: true}.equal(HMA{wma: WMA{}}))
	assert.False(t, HMA{wma: WMA{length: 2}}.equal(HMA{wma: WMA{}}))
	assert.False(t, HMA{wma: WMA{}}.equal(WMA{}))
}

func Test_HMA_WMA(t *testing.T) {
	assert.Equal(t, WMA{length: 1, offset: 2}, HMA{wma: WMA{length: 1, offset: 2}}.WMA())
}

func Test_HMA_Offset(t *testing.T) {
	assert.Equal(t, 3, HMA{wma: WMA{offset: 3}}.Offset())
}

func Test_HMA_validate(t *testing.T) {
	cc := map[string]struct {
		HMA   HMA
		Error error
		Valid bool
	}{
		"Invalid WMA": {
			HMA:   HMA{wma: WMA{length: -1, offset: 2}},
			Error: assert.AnError,
			Valid: false,
		},
		"Successful validation": {
			HMA:   HMA{wma: WMA{length: 2, offset: 2}},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.HMA.validate())
			assert.Equal(t, c.Valid, c.HMA.valid)
		})
	}
}

func Test_HMA_Calc(t *testing.T) {
	cc := map[string]struct {
		HMA    HMA
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			HMA:   HMA{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			HMA: HMA{wma: WMA{length: 5, offset: 0, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation": {
			HMA: HMA{wma: WMA{length: 3, offset: 0, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
				decimal.NewFromInt(30),
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
			},
			Result: decimal.RequireFromString("31.5"),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.HMA.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_HMA_Count(t *testing.T) {
	assert.Equal(t, 31, HMA{wma: WMA{length: 15, offset: 2}}.Count())
}

func Test_HMA_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result HMA
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/}`,
			Error: assert.AnError,
		},
		"Invalid creation": {
			JSON:  `{"wma":{"length":-1}}`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"wma":{"length":1}}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"wma":{"length":3,"offset":3}}`,
			Result: HMA{wma: WMA{length: 3, offset: 3, valid: true}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := HMA{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_HMA_MarshalJSON(t *testing.T) {
	d, err := HMA{wma: WMA{length: 3, offset: 2}}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"wma":{"length":3,"offset":2}}`, string(d))
}

func Test_HMA_namedMarshalJSON(t *testing.T) {
	d, err := HMA{wma: WMA{length: 3, offset: 1}}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"hma","wma":{"length":3,"offset":1}}`, string(d))
}

func Test_NewCD(t *testing.T) {
	cc := map[string]struct {
		Percent bool
		Source1 Indicator
		Source2 Indicator
		Offset  int
		Result  CD
		Error   error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Percent: true,
			Source1: &IndicatorMock{},
			Source2: &IndicatorMock{},
			Offset:  4,
			Result:  CD{percent: true, source1: &IndicatorMock{}, source2: &IndicatorMock{}, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewCD(c.Percent, c.Source1, c.Source2, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_CD_Equal(t *testing.T) {
	assert.True(t, CD{source1: SMA{}, source2: SMA{}}.equal(CD{source1: SMA{}, source2: SMA{}}))
	assert.False(t, CD{source1: SMA{}, source2: SMA{}, offset: 2}.equal(CD{source1: SMA{}, source2: SMA{}}))
	assert.False(t, CD{source1: SMA{length: 1}, source2: SMA{}}.equal(CD{source1: SMA{}, source2: SMA{}}))
	assert.False(t, CD{source1: SMA{}, source2: SMA{length: 1}}.equal(CD{source1: SMA{}, source2: SMA{}}))
	assert.False(t, CD{source1: SMA{}, source2: SMA{}}.equal(SMA{}))
}

func Test_CD_Percent(t *testing.T) {
	assert.Equal(t, true, CD{percent: true}.Percent())
}

func Test_CD_Sub1(t *testing.T) {
	assert.Equal(t, &IndicatorMock{}, CD{source1: &IndicatorMock{}, source2: nil}.Sub1())
}

func Test_CD_Sub2(t *testing.T) {
	assert.Equal(t, &IndicatorMock{}, CD{source1: nil, source2: &IndicatorMock{}}.Sub2())
}

func Test_CD_Offset(t *testing.T) {
	assert.Equal(t, 4, CD{offset: 4}.Offset())
}

func Test_CD_validate(t *testing.T) {
	cc := map[string]struct {
		CD    CD
		Error error
		Valid bool
	}{
		"Invalid source1": {
			CD:    CD{source1: nil, source2: &IndicatorMock{}, offset: 4},
			Error: ErrInvalidSource,
			Valid: false,
		},
		"Invalid source2": {
			CD:    CD{source1: &IndicatorMock{}, source2: nil, offset: 4},
			Error: ErrInvalidSource,
			Valid: false,
		},
		"Invalid offset": {
			CD:    CD{source1: &IndicatorMock{}, source2: &IndicatorMock{}, offset: -4},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			CD:    CD{source1: &IndicatorMock{}, source2: &IndicatorMock{}, offset: 4},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.CD.validate())
			assert.Equal(t, c.Valid, c.CD.valid)
		})
	}
}

func Test_CD_Calc(t *testing.T) {
	stubIndicator := func(v decimal.Decimal, e error, a int) *IndicatorMock {
		return &IndicatorMock{
			CalcFunc: func(dd []decimal.Decimal) (decimal.Decimal, error) {
				return v, e
			},
			CountFunc: func() int {
				return a
			},
		}
	}

	cc := map[string]struct {
		CD     CD
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			CD:    CD{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size for source1": {
			CD: CD{source1: stubIndicator(decimal.Zero, nil, 10), source2: stubIndicator(decimal.Zero, nil, 1), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Invalid data size for source2": {
			CD: CD{source1: stubIndicator(decimal.Zero, nil, 1), source2: stubIndicator(decimal.Zero, nil, 10), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Invalid source1": {
			CD: CD{source1: stubIndicator(decimal.Zero, assert.AnError, 1), source2: stubIndicator(decimal.Zero, nil, 1), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: assert.AnError,
		},
		"Invalid source2": {
			CD: CD{source1: stubIndicator(decimal.Zero, nil, 1), source2: stubIndicator(decimal.Zero, assert.AnError, 1), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: assert.AnError,
		},
		"Successful calculation with offset": {
			CD: CD{source1: stubIndicator(decimal.NewFromInt(5), nil, 1), source2: stubIndicator(decimal.NewFromInt(10), nil, 1), offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(5),
		},
		"Successful calculation using percent": {
			CD: CD{percent: true, source1: stubIndicator(decimal.NewFromInt(5), nil, 1), source2: stubIndicator(decimal.NewFromInt(10), nil, 1), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(100),
		},
		"Successful calculation": {
			CD: CD{source1: stubIndicator(decimal.NewFromInt(5), nil, 1), source2: stubIndicator(decimal.NewFromInt(10), nil, 1), offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(5),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.CD.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_CD_Count(t *testing.T) {
	stubIndicator := func(a int) *IndicatorMock {
		return &IndicatorMock{
			CountFunc: func() int {
				return a
			},
		}
	}

	assert.Equal(t, 15, CD{source1: stubIndicator(15), source2: stubIndicator(10)}.Count())
}

func Test_CD_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result CD
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\-_-/}`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"source1":{"name":"sma","length":1},"source2":{"name":"sma","length":1},"offset":-2}`,
			Error: assert.AnError,
		},
		"Invalid source1": {
			JSON:  `{"source2":{"name":"sma","length":1}}`,
			Error: assert.AnError,
		},
		"Invalid source2": {
			JSON:  `{"source1":{"name":"sma","length":1}}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"percent":true,"source1":{"name":"sma","length":1,"offset":4},"source2":{"name":"sma","length":2,"offset":6},"offset":5}`,
			Result: CD{percent: true, source1: SMA{length: 1, offset: 4, valid: true}, source2: SMA{length: 2, offset: 6, valid: true}, offset: 5, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := CD{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_CD_MarshalJSON(t *testing.T) {
	stubIndicator := func(d []byte, e error) *IndicatorMock {
		return &IndicatorMock{
			namedMarshalJSONFunc: func() ([]byte, error) {
				return d, e
			},
		}
	}

	cc := map[string]struct {
		CD     CD
		Result string
		Error  error
	}{
		"Error returned during source1 marshalling": {
			CD:    CD{source1: stubIndicator(nil, assert.AnError), source2: stubIndicator([]byte(`{"name":"indicatormock"}`), nil)},
			Error: assert.AnError,
		},
		"Error returned during source2 marshalling": {
			CD:    CD{source1: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), source2: stubIndicator(nil, assert.AnError)},
			Error: assert.AnError,
		},
		"Successful marshal": {
			CD:     CD{percent: true, source1: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), source2: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), offset: 4},
			Result: `{"percent":true,"source1":{"name":"indicatormock"},"source2":{"name":"indicatormock"},"offset":4}`,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			d, err := c.CD.MarshalJSON()
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.JSONEq(t, c.Result, string(d))
		})
	}
}

func Test_CD_namedMarshalJSON(t *testing.T) {
	stubIndicator := func(d []byte, e error) *IndicatorMock {
		return &IndicatorMock{
			namedMarshalJSONFunc: func() ([]byte, error) {
				return d, e
			},
		}
	}

	cc := map[string]struct {
		CD     CD
		Result string
		Error  error
	}{
		"Error returned during source1 marshalling": {
			CD:    CD{source1: stubIndicator(nil, assert.AnError), source2: stubIndicator([]byte(`{"name":"indicatormock"}`), nil)},
			Error: assert.AnError,
		},
		"Error returned during source2 marshalling": {
			CD:    CD{source1: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), source2: stubIndicator(nil, assert.AnError)},
			Error: assert.AnError,
		},
		"Successful marshal": {
			CD:     CD{percent: true, source1: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), source2: stubIndicator([]byte(`{"name":"indicatormock"}`), nil), offset: 4},
			Result: `{"name":"cd","percent":true,"source1":{"name":"indicatormock"},"source2":{"name":"indicatormock"},"offset":4}`,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			d, err := c.CD.namedMarshalJSON()
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.JSONEq(t, c.Result, string(d))
		})
	}
}

func Test_NewROC(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result ROC
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 4,
			Result: ROC{length: 1, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewROC(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_ROC_Equal(t *testing.T) {
	assert.True(t, ROC{}.equal(ROC{}))
	assert.False(t, ROC{length: 1}.equal(ROC{}))
	assert.False(t, ROC{}.equal(RSI{}))
}

func Test_ROC_Length(t *testing.T) {
	assert.Equal(t, 1, ROC{length: 1}.Length())
}

func Test_ROC_Offset(t *testing.T) {
	assert.Equal(t, 1, ROC{offset: 1}.Offset())
}

func Test_ROC_validate(t *testing.T) {
	cc := map[string]struct {
		ROC   ROC
		Error error
		Valid bool
	}{
		"Invalid length": {
			ROC:   ROC{length: -1, offset: 2},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			ROC:   ROC{length: 2, offset: -1},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			ROC:   ROC{length: 1, offset: 1},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.ROC.validate())
			assert.Equal(t, c.Valid, c.ROC.valid)
		})
	}
}

func Test_ROC_Calc(t *testing.T) {
	cc := map[string]struct {
		ROC    ROC
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			ROC:   ROC{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			ROC: ROC{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful handled division by 0": {
			ROC: ROC{length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(420),
				decimal.NewFromInt(0),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
			},
			Result: decimal.Zero,
		},
		"Successful calculation with offset": {
			ROC: ROC{length: 5, offset: 3, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(7),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(10),
				decimal.NewFromInt(11),
				decimal.NewFromInt(12),
				decimal.NewFromInt(10),
			},
			Result: decimal.RequireFromString("42.85714285714286"),
		},
		"Successful calculation": {
			ROC: ROC{length: 5, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(7),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(10),
			},
			Result: decimal.RequireFromString("42.85714285714286"),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.ROC.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_ROC_Count(t *testing.T) {
	assert.Equal(t, 17, ROC{length: 15, offset: 2}.Count())
}

func Test_ROC_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result ROC
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":2}`,
			Result: ROC{length: 1, offset: 2, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := ROC{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_ROC_MarshalJSON(t *testing.T) {
	d, err := ROC{length: 1, offset: 2}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":2}`, string(d))
}

func Test_ROC_namedMarshalJSON(t *testing.T) {
	d, err := ROC{length: 1, offset: 3}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"roc","length":1,"offset":3}`, string(d))
}

func Test_NewRSI(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result RSI
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 4,
			Result: RSI{length: 1, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewRSI(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_RSI_Equal(t *testing.T) {
	assert.True(t, RSI{}.equal(RSI{}))
	assert.False(t, RSI{length: 1}.equal(RSI{}))
	assert.False(t, RSI{}.equal(ROC{}))
}

func Test_RSI_Length(t *testing.T) {
	assert.Equal(t, 1, RSI{length: 1}.Length())
}

func Test_RSI_Offset(t *testing.T) {
	assert.Equal(t, 4, RSI{offset: 4}.Offset())
}

func Test_RSI_validate(t *testing.T) {
	cc := map[string]struct {
		RSI   RSI
		Error error
		Valid bool
	}{
		"Invalid length": {
			RSI:   RSI{length: 0, offset: 4},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			RSI:   RSI{length: 2, offset: -4},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			RSI:   RSI{length: 1, offset: 0},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.RSI.validate())
			assert.Equal(t, c.Valid, c.RSI.valid)
		})
	}
}

func Test_RSI_Calc(t *testing.T) {
	cc := map[string]struct {
		RSI    RSI
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			RSI:   RSI{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			RSI: RSI{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation when average gain 0": {
			RSI: RSI{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(16),
				decimal.NewFromInt(12),
				decimal.NewFromInt(8),
			},
			Result: decimal.NewFromInt(0),
		},
		"Successful calculation when average loss 0": {
			RSI: RSI{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(2),
				decimal.NewFromInt(4),
				decimal.NewFromInt(8),
			},
			Result: Hundred,
		},
		"Successful calculation with offset": {
			RSI: RSI{length: 3, offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(8),
				decimal.NewFromInt(12),
				decimal.NewFromInt(8),
				decimal.NewFromInt(58),
				decimal.NewFromInt(58),
			},
			Result: decimal.NewFromInt(50),
		},
		"Successful calculation": {
			RSI: RSI{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(8),
				decimal.NewFromInt(12),
				decimal.NewFromInt(8),
			},
			Result: decimal.NewFromInt(50),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.RSI.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_RSI_Count(t *testing.T) {
	assert.Equal(t, 15, RSI{length: 15}.Count())
}

func Test_RSI_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result RSI
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":2}`,
			Result: RSI{length: 1, offset: 2, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := RSI{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_RSI_MarshalJSON(t *testing.T) {
	d, err := RSI{length: 1, offset: 4}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":4}`, string(d))
}

func Test_RSI_namedMarshalJSON(t *testing.T) {
	d, err := RSI{length: 1, offset: 2}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"rsi","length":1,"offset":2}`, string(d))
}

func Test_NewSMA(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result SMA
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 3,
			Result: SMA{length: 1, offset: 3, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewSMA(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_SMA_Equal(t *testing.T) {
	assert.True(t, SMA{}.equal(SMA{}))
	assert.False(t, SMA{length: 1}.equal(SMA{}))
	assert.False(t, SMA{}.equal(ROC{}))
}

func Test_SMA_Length(t *testing.T) {
	assert.Equal(t, 1, SMA{length: 1}.Length())
}

func Test_SMA_Offset(t *testing.T) {
	assert.Equal(t, 4, SMA{offset: 4}.Offset())
}

func Test_SMA_validate(t *testing.T) {
	cc := map[string]struct {
		SMA   SMA
		Error error
		Valid bool
	}{
		"Invalid length": {
			SMA:   SMA{length: 0, offset: 2},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			SMA:   SMA{length: 2, offset: -2},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			SMA:   SMA{length: 1, offset: 0},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.SMA.validate())
			assert.Equal(t, c.Valid, c.SMA.valid)
		})
	}
}

func Test_SMA_Calc(t *testing.T) {
	cc := map[string]struct {
		SMA    SMA
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			SMA:   SMA{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			SMA: SMA{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with offset": {
			SMA: SMA{length: 3, offset: 2, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
				decimal.NewFromInt(2),
				decimal.NewFromInt(2),
			},
			Result: decimal.NewFromInt(31),
		},
		"Successful calculation": {
			SMA: SMA{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
				decimal.NewFromInt(31),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(31),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.SMA.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_SMA_Count(t *testing.T) {
	assert.Equal(t, 18, SMA{length: 15, offset: 3}.Count())
}

func Test_SMA_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result SMA
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":4}`,
			Result: SMA{length: 1, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := SMA{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_SMA_MarshalJSON(t *testing.T) {
	d, err := SMA{length: 1, offset: 3}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":3}`, string(d))
}

func Test_SMA_namedMarshalJSON(t *testing.T) {
	d, err := SMA{length: 1, offset: 4}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"sma","length":1,"offset":4}`, string(d))
}

func Test_NewSRSI(t *testing.T) {
	cc := map[string]struct {
		RSI    RSI
		Result SRSI
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			RSI:    RSI{length: 1, offset: 3, valid: true},
			Result: SRSI{rsi: RSI{length: 1, offset: 3, valid: true}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewSRSI(c.RSI)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_SRSI_Equal(t *testing.T) {
	assert.True(t, SRSI{}.equal(SRSI{}))
	assert.False(t, SRSI{valid: true}.equal(SRSI{}))
	assert.False(t, SRSI{rsi: RSI{length: 1}}.equal(SRSI{}))
	assert.False(t, SRSI{}.equal(ROC{}))
}

func Test_SRSI_RSI(t *testing.T) {
	assert.Equal(t, RSI{length: 1}, SRSI{rsi: RSI{length: 1}}.RSI())
}

func Test_SRSI_Offset(t *testing.T) {
	assert.Equal(t, 3, SRSI{rsi: RSI{offset: 3}}.Offset())
}

func Test_SRSI_validate(t *testing.T) {
	cc := map[string]struct {
		SRSI  SRSI
		Error error
		Valid bool
	}{
		"Invalid RSI": {
			SRSI:  SRSI{},
			Error: assert.AnError,
			Valid: false,
		},
		"Successful validation": {
			SRSI:  SRSI{rsi: RSI{length: 1}},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.SRSI.validate())
			assert.Equal(t, c.Valid, c.SRSI.valid)
		})
	}
}

func Test_SRSI_Calc(t *testing.T) {
	cc := map[string]struct {
		SRSI   SRSI
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			SRSI:  SRSI{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			SRSI: SRSI{rsi: RSI{length: 5, offset: 0, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: assert.AnError,
		},
		"Successfully handled division by 0": {
			SRSI: SRSI{rsi: RSI{length: 3, offset: 0, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(8),
				decimal.NewFromInt(12),
				decimal.NewFromInt(8),
				decimal.NewFromInt(12),
				decimal.NewFromInt(8),
			},
			Result: decimal.Zero,
		},
		"Successful calculation with offset": {
			SRSI: SRSI{rsi: RSI{length: 3, offset: 3, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(4),
				decimal.NewFromInt(10),
				decimal.NewFromInt(6),
				decimal.NewFromInt(8),
				decimal.NewFromInt(6),
				decimal.NewFromInt(64),
				decimal.NewFromInt(84),
				decimal.NewFromInt(64),
			},
			Result: decimal.RequireFromString("0.625"),
		},
		"Successful calculation": {
			SRSI: SRSI{rsi: RSI{length: 3, offset: 0, valid: true}, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(4),
				decimal.NewFromInt(10),
				decimal.NewFromInt(6),
				decimal.NewFromInt(8),
				decimal.NewFromInt(6),
			},
			Result: decimal.RequireFromString("0.625"),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.SRSI.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_SRSI_Count(t *testing.T) {
	assert.Equal(t, 32, SRSI{rsi: RSI{length: 15, offset: 3}}.Count())
}

func Test_SRSI_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result SRSI
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"rsi":{"length":1,"offset":2}}`,
			Result: SRSI{rsi: RSI{length: 1, offset: 2, valid: true}, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := SRSI{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_SRSI_MarshalJSON(t *testing.T) {
	d, err := SRSI{rsi: RSI{length: 1, offset: 2}}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"rsi":{"length":1,"offset":2}}`, string(d))
}

func Test_SRSI_namedMarshalJSON(t *testing.T) {
	d, err := SRSI{rsi: RSI{length: 1, offset: 2}}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.Equal(t, `{"name":"srsi","rsi":{"length":1,"offset":2}}`, string(d))
}

func Test_NewStoch(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result Stoch
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 4,
			Result: Stoch{length: 1, offset: 4, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewStoch(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_Stoch_Equal(t *testing.T) {
	assert.True(t, Stoch{}.equal(Stoch{}))
	assert.False(t, Stoch{length: 1}.equal(Stoch{}))
	assert.False(t, Stoch{}.equal(ROC{}))
}

func Test_Stoch_Length(t *testing.T) {
	assert.Equal(t, 1, Stoch{length: 1}.Length())
}

func Test_Stoch_Offset(t *testing.T) {
	assert.Equal(t, 3, Stoch{offset: 3}.Offset())
}

func Test_Stoch_validate(t *testing.T) {
	cc := map[string]struct {
		Stoch Stoch
		Error error
		Valid bool
	}{
		"Invalid length": {
			Stoch: Stoch{length: 0, offset: 0},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			Stoch: Stoch{length: 2, offset: -1},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			Stoch: Stoch{length: 1, offset: 2},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.Stoch.validate())
			assert.Equal(t, c.Valid, c.Stoch.valid)
		})
	}
}

func Test_Stoch_Calc(t *testing.T) {
	cc := map[string]struct {
		Stoch  Stoch
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			Stoch: Stoch{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			Stoch: Stoch{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation when new lows are reached": {
			Stoch: Stoch{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(150),
				decimal.NewFromInt(125),
				decimal.NewFromInt(145),
			},
			Result: decimal.NewFromInt(80),
		},
		"Successfully handled division by 0": {
			Stoch: Stoch{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(150),
				decimal.NewFromInt(150),
				decimal.NewFromInt(150),
			},
			Result: decimal.Zero,
		},
		"Successful calculation when new highs are reached": {
			Stoch: Stoch{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(120),
				decimal.NewFromInt(145),
				decimal.NewFromInt(135),
			},
			Result: decimal.NewFromInt(60),
		},
		"Successful calculation with offset": {
			Stoch: Stoch{length: 3, offset: 3, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(120),
				decimal.NewFromInt(145),
				decimal.NewFromInt(135),
				decimal.NewFromInt(350),
				decimal.NewFromInt(300),
				decimal.NewFromInt(420),
			},
			Result: decimal.NewFromInt(60),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.Stoch.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_Stoch_Count(t *testing.T) {
	assert.Equal(t, 18, Stoch{length: 15, offset: 3}.Count())
}

func Test_Stoch_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result Stoch
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{"length": "down"}`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0,"offset":3}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":3}`,
			Result: Stoch{length: 1, offset: 3, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := Stoch{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_Stoch_MarshalJSON(t *testing.T) {
	d, err := Stoch{length: 1, offset: 3}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":3}`, string(d))
}

func Test_Stoch_namedMarshalJSON(t *testing.T) {
	d, err := Stoch{length: 1, offset: 2}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"stoch","length":1,"offset":2}`, string(d))
}

func Test_NewWMA(t *testing.T) {
	cc := map[string]struct {
		Length int
		Offset int
		Result WMA
		Error  error
	}{
		"Invalid parameters": {
			Error: assert.AnError,
		},
		"Successful creation": {
			Length: 1,
			Offset: 3,
			Result: WMA{length: 1, offset: 3, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v, err := NewWMA(c.Length, c.Offset)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_WMA_Equal(t *testing.T) {
	assert.True(t, WMA{}.equal(WMA{}))
	assert.False(t, WMA{length: 1}.equal(WMA{}))
	assert.False(t, WMA{}.equal(ROC{}))
}

func Test_WMA_Length(t *testing.T) {
	assert.Equal(t, 1, WMA{length: 1}.Length())
}

func Test_WMA_Offset(t *testing.T) {
	assert.Equal(t, 2, WMA{offset: 2}.Offset())
}

func Test_WMA_validate(t *testing.T) {
	cc := map[string]struct {
		WMA   WMA
		Error error
		Valid bool
	}{
		"Invalid length": {
			WMA:   WMA{length: 0, offset: 0},
			Error: ErrInvalidLength,
			Valid: false,
		},
		"Invalid offset": {
			WMA:   WMA{length: 2, offset: -1},
			Error: ErrInvalidOffset,
			Valid: false,
		},
		"Successful validation": {
			WMA:   WMA{length: 1, offset: 3},
			Valid: true,
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			equalError(t, c.Error, c.WMA.validate())
			assert.Equal(t, c.Valid, c.WMA.valid)
		})
	}
}

func Test_WMA_Calc(t *testing.T) {
	cc := map[string]struct {
		WMA    WMA
		Data   []decimal.Decimal
		Result decimal.Decimal
		Error  error
	}{
		"Invalid indicator": {
			WMA:   WMA{},
			Error: ErrInvalidIndicator,
		},
		"Invalid data size": {
			WMA: WMA{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(30),
			},
			Error: ErrInvalidDataSize,
		},
		"Successful calculation with offset": {
			WMA: WMA{length: 3, offset: 3, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(30),
				decimal.NewFromInt(30),
				decimal.NewFromInt(32),
				decimal.NewFromInt(320),
				decimal.NewFromInt(33),
				decimal.NewFromInt(325),
			},
			Result: decimal.NewFromInt(31),
		},
		"Successful calculation": {
			WMA: WMA{length: 3, offset: 0, valid: true},
			Data: []decimal.Decimal{
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(420),
				decimal.NewFromInt(30),
				decimal.NewFromInt(30),
				decimal.NewFromInt(32),
			},
			Result: decimal.NewFromInt(31),
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			res, err := c.WMA.Calc(c.Data)
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result.String(), res.String())
		})
	}
}

func Test_WMA_Count(t *testing.T) {
	assert.Equal(t, 18, WMA{length: 15, offset: 3}.Count())
}

func Test_WMA_UnmarshalJSON(t *testing.T) {
	cc := map[string]struct {
		JSON   string
		Result WMA
		Error  error
	}{
		"Invalid JSON": {
			JSON:  `{\"_"/`,
			Error: assert.AnError,
		},
		"Invalid validation": {
			JSON:  `{"length":0,"offset":0}`,
			Error: assert.AnError,
		},
		"Successful unmarshal": {
			JSON:   `{"length":1,"offset":2}`,
			Result: WMA{length: 1, offset: 2, valid: true},
		},
	}

	for cn, c := range cc {
		c := c

		t.Run(cn, func(t *testing.T) {
			t.Parallel()

			v := WMA{}
			err := v.UnmarshalJSON([]byte(c.JSON))
			equalError(t, c.Error, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.Result, v)
		})
	}
}

func Test_WMA_MarshalJSON(t *testing.T) {
	d, err := WMA{length: 1, offset: 3}.MarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"length":1,"offset":3}`, string(d))
}

func Test_WMA_namedMarshalJSON(t *testing.T) {
	d, err := WMA{length: 1, offset: 2}.namedMarshalJSON()

	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"wma","length":1, "offset":2}`, string(d))
}
