package indc

import (
	"encoding/json"
	"errors"
	"math"

	"github.com/shopspring/decimal"
)

// Indicator is an interface that every indicator should implement.
//go:generate moq -out ./mock_test.go . Indicator
type Indicator interface {
	// Calc should calculate indicator's value.
	Calc(dd []decimal.Decimal) (decimal.Decimal, error)

	// Count should determine the total amount of data points needed
	// for indicator's calculation.
	Count() int

	// Offset should determine how many data points should be skipped
	// from the end during the calculations.
	Offset() int

	// namedMarshalJSON converts indicator and its name to JSON.
	namedMarshalJSON() ([]byte, error)

	equal(i Indicator) bool
}

// NameAroon returns Aroon indicator name.
const NameAroon = "aroon"

// Aroon holds all the necessary information needed to calculate Aroon.
// The zero value is not usable.
type Aroon struct {
	// valid specifies whether Aroon paremeters were validated.
	valid bool

	// trend specifies which Aroon trend to use during the
	// calculation process. Allowed values: up, down.
	trend String

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewAroon validates provided configuration options and
// creates new Aroon indicator.
func NewAroon(trend String, length, offset int) (Aroon, error) {
	a := Aroon{trend: trend, length: length, offset: offset}

	if err := a.validate(); err != nil {
		return Aroon{}, err
	}

	return a, nil
}

// Equal checks whether provided aroon has exactly the same values as main
// aroon.
func (a Aroon) Equal(a1 Aroon) bool {
	return a == a1
}

func (a Aroon) equal(i Indicator) bool {
	a1, ok := i.(Aroon)
	if ok {
		return a.Equal(a1)
	}

	return ok
}

// Trend returns trend configuration option.
func (a Aroon) Trend() String {
	return a.trend
}

// Length returns length configuration option.
func (a Aroon) Length() int {
	return a.length
}

// Offset returns offset configuration option.
func (a Aroon) Offset() int {
	return a.offset
}

// validate checks whether Aroon was configured properly or not.
func (a *Aroon) validate() error {
	if a.trend != "down" && a.trend != "up" {
		return errors.New("invalid trend")
	}

	if a.length < 1 {
		return ErrInvalidLength
	}

	if a.offset < 0 {
		return ErrInvalidOffset
	}

	a.valid = true

	return nil
}

// Calc calculates Aroon from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/a/aroon.asp.
// All credits are due to Tushar Chande who developed Aroon indicator.
func (a Aroon) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !a.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, a.Count()-a.offset, a.offset)
	if err != nil {
		return decimal.Zero, err
	}

	v := decimal.Zero
	p := decimal.Zero

	for i := 0; i < len(dd); i++ {
		if v.Equal(decimal.Zero) {
			v = dd[i]
		}

		if a.trend == "up" && v.LessThanOrEqual(dd[i]) ||
			a.trend == "down" && !v.LessThan(dd[i]) {

			v = dd[i]
			p = decimal.NewFromInt(int64(a.length - i - 1))
		}
	}

	return decimal.NewFromInt(int64(a.length)).Sub(p).
		Mul(Hundred).Div(decimal.NewFromInt(int64(a.length))), nil
}

// Count determines the total amount of data points needed for Aroon
// calculation.
func (a Aroon) Count() int {
	return a.length + a.offset
}

// UnmarshalJSON parses JSON into Aroon structure.
func (a *Aroon) UnmarshalJSON(d []byte) error {
	var data struct {
		Trend  String `json:"trend"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	na, err := NewAroon(data.Trend, data.Length, data.Offset)
	if err != nil {
		return err
	}

	*a = na

	return nil
}

// MarshalJSON converts Aroon configuration data into JSON.
func (a Aroon) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Trend  String `json:"trend"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Trend:  a.trend,
		Length: a.length,
		Offset: a.offset,
	})
}

// namedMarshalJSON converts Aroon configuration data with its
// name into JSON.
func (a Aroon) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Trend  String `json:"trend"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameAroon,
		Trend:  a.trend,
		Length: a.length,
		Offset: a.offset,
	})
}

// NameCCI returns CCI indicator name.
const NameCCI = "cci"

// CCI holds all the necessary information needed to calculate commodity
// channel index.
// The zero value is not usable.
type CCI struct {
	// valid specifies whether CCI paremeters were validated.
	valid bool

	// source specifies which indicator to use during calculation process.
	source Indicator

	// factor is used to scale CCI to provide more readable numbers.
	// default is 0.015f.
	factor decimal.Decimal
}

// NewCCI validates provided configuration options and creates
// new CCI indicator.
// If provided factor is zero, default value is going to be used (0.015f).
func NewCCI(source Indicator, factor decimal.Decimal) (CCI, error) {
	if factor.Equal(decimal.Zero) {
		factor = decimal.RequireFromString("0.015")
	}

	c := CCI{source: source, factor: factor}

	if err := c.validate(); err != nil {
		return CCI{}, err
	}

	return c, nil
}

// Equal checks whether provided cci has exactly the same values as main
// cci.
func (c CCI) Equal(c1 CCI) bool {
	if c.valid != c1.valid || !c.factor.Equal(c1.factor) {
		return false
	}

	return c.source.equal(c1.source)
}

func (c CCI) equal(i Indicator) bool {
	c1, ok := i.(CCI)
	if ok {
		return c.Equal(c1)
	}

	return ok
}

// Sub returns source configuration option.
func (c CCI) Sub() Indicator {
	return c.source
}

// Factor returns factor configuration option.
func (c CCI) Factor() decimal.Decimal {
	return c.factor
}

// Offset returns Offset configuration option.
func (c CCI) Offset() int {
	return c.source.Offset()
}

// validate checks whether CCI was configured properly or not.
func (c *CCI) validate() error {
	if c.source == nil {
		return ErrInvalidSource
	}

	if c.factor.LessThanOrEqual(decimal.Zero) {
		return errors.New("invalid factor")
	}

	c.valid = true

	return nil
}

// Calc calculates CCI from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/c/commoditychannelindex.asp.
// All credits are due to Donald Lambert who developed CCI indicator.
func (c CCI) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !c.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, c.Count(), 0)
	if err != nil {
		return decimal.Zero, err
	}

	m, err := c.source.Calc(dd)
	if err != nil {
		return decimal.Zero, err
	}

	denom := c.factor.Mul(meanDeviation(dd))

	if denom.Equal(decimal.Zero) {
		return decimal.Zero, nil
	}

	return dd[len(dd)-1].Sub(m).Div(denom), nil
}

// Count determines the total amount of data points needed for CCI
// calculation.
func (c CCI) Count() int {
	return c.source.Count()
}

// UnmarshalJSON parses JSON into CCI structure.
func (c *CCI) UnmarshalJSON(d []byte) error {
	var data struct {
		Source json.RawMessage `json:"source"`
		Factor string          `json:"factor"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	src, err := fromJSON(data.Source)
	if err != nil {
		return err
	}

	if data.Factor == "" {
		data.Factor = "0"
	}

	fac, err := decimal.NewFromString(data.Factor)
	if err != nil {
		return err
	}

	cn, err := NewCCI(src, fac)
	if err != nil {
		return err
	}

	*c = cn

	return nil
}

// MarshalJSON converts CCI configuration data into JSON.
func (c CCI) MarshalJSON() ([]byte, error) {
	src, err := c.source.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Source json.RawMessage `json:"source"`
		Factor string          `json:"factor"`
	}{
		Source: src,
		Factor: c.factor.String(),
	})
}

// namedMarshalJSON converts CCI configuration data with its
// name into JSON.
func (c CCI) namedMarshalJSON() ([]byte, error) {
	src, err := c.source.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Name   String          `json:"name"`
		Source json.RawMessage `json:"source"`
		Factor string          `json:"factor"`
	}{
		Name:   NameCCI,
		Source: src,
		Factor: c.factor.String(),
	})
}

// NameDEMA returns DEMA indicator name.
const NameDEMA = "dema"

// DEMA holds all the necessary information needed to calculate
// double exponential moving average.
// The zero value is not usable.
type DEMA struct {
	// valid specifies whether DEMA paremeters were validated.
	valid bool

	// ema specifies what ema should be used for dema calculations.
	ema EMA
}

// NewDEMA validates provided configuration options and creates
// new DEMA indicator.
func NewDEMA(ema EMA) (DEMA, error) {
	dm := DEMA{ema: ema}

	if err := dm.validate(); err != nil {
		return DEMA{}, err
	}

	return dm, nil
}

// Equal checks whether provided dema has exactly the same values as main
// dema.
func (dm DEMA) Equal(dm1 DEMA) bool {
	if dm.valid != dm1.valid {
		return false
	}

	return dm.ema.Equal(dm1.ema)
}

func (dm DEMA) equal(i Indicator) bool {
	dm1, ok := i.(DEMA)
	if ok {
		return dm.Equal(dm1)
	}

	return ok
}

// EMA returns ema configuration option.
func (dm DEMA) EMA() EMA {
	return dm.ema
}

// Offset returns offset configuration option.
func (dm DEMA) Offset() int {
	return dm.ema.offset
}

// validate checks whether DEMA was configured properly or not.
func (dm *DEMA) validate() error {
	if err := dm.ema.validate(); err != nil {
		return err
	}

	dm.valid = true

	return nil
}

// Calc calculates DEMA from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/d/double-exponential-moving-average.asp.
// All credits are due to Patrick Mulloy who developed DEMA indicator.
func (dm DEMA) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !dm.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, dm.Count()-dm.ema.offset, dm.ema.offset)
	if err != nil {
		return decimal.Zero, err
	}

	v := make([]decimal.Decimal, dm.ema.Length())

	s, _ := NewSMA(dm.ema.length, 0)
	v[0], _ = s.Calc(dd[:dm.ema.Length()])

	for i := dm.ema.Length(); i < len(dd); i++ {
		v[i-dm.ema.Length()+1], _ = dm.ema.CalcNext(v[i-dm.ema.Length()], dd[i])
	}

	r := v[0]

	for i := 0; i < len(v); i++ {
		r, _ = dm.ema.CalcNext(r, v[i])
	}

	return r, nil
}

// Count determines the total amount of data points needed for DEMA
// calculation.
func (dm DEMA) Count() int {
	return dm.ema.Count()
}

// UnmarshalJSON parses JSON into DEMA structure.
func (dm *DEMA) UnmarshalJSON(d []byte) error {
	var data struct {
		EMA struct {
			Length int `json:"length"`
			Offset int `json:"offset"`
		} `json:"ema"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	ne, err := NewEMA(data.EMA.Length, data.EMA.Offset)
	if err != nil {
		return err
	}

	ndm, _ := NewDEMA(ne)

	*dm = ndm

	return nil
}

// MarshalJSON converts DEMA configuration data into JSON.
func (dm DEMA) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EMA EMA `json:"ema"`
	}{
		EMA: dm.ema,
	})
}

// namedMarshalJSON converts DEMA configuration data with its
// name into JSON.
func (dm DEMA) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name String `json:"name"`
		EMA  EMA    `json:"ema"`
	}{
		Name: NameDEMA,
		EMA:  dm.ema,
	})
}

// NameEMA returns EMA indicator name.
const NameEMA = "ema"

// EMA holds all the necessary information needed to calculate exponential
// moving average.
// The zero value is not usable.
type EMA struct {
	SMA
}

// NewEMA validates provided configuration options and
// creates new EMA indicator.
func NewEMA(length, offset int) (EMA, error) {
	s, err := NewSMA(length, offset)
	if err != nil {
		return EMA{}, err
	}

	e := EMA{SMA: s}

	return e, nil
}

// Equal checks whether provided ema has exactly the same values as main
// ema.
func (e EMA) Equal(e1 EMA) bool {
	return e.SMA.Equal(e1.SMA)
}

func (e EMA) equal(i Indicator) bool {
	e1, ok := i.(EMA)
	if ok {
		return e.Equal(e1)
	}

	return ok
}

// Calc calculates EMA from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/e/ema.asp.
func (e EMA) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !e.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, e.Count()-e.offset, e.offset)
	if err != nil {
		return decimal.Zero, err
	}

	s, _ := NewSMA(e.length, 0)
	r, _ := s.Calc(dd[:e.length])

	for i := e.length; i < len(dd); i++ {
		r, _ = e.CalcNext(r, dd[i])
	}

	return r, nil
}

// CalcNext calculates sequential EMA by using previous EMA.
func (e EMA) CalcNext(l, n decimal.Decimal) (decimal.Decimal, error) {
	if !e.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	m := e.multiplier()

	return n.Mul(m).Add(l.Mul(decimal.NewFromInt(1).Sub(m))), nil
}

// multiplier calculates EMA multiplier.
func (e EMA) multiplier() decimal.Decimal {
	return decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(e.Length()) + 1))
}

// Count determines the total amount of data points needed for EMA
// calculation.
func (e EMA) Count() int {
	return e.length*2 + e.offset - 1
}

// UnmarshalJSON parses JSON into EMA structure.
func (e *EMA) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	ne, err := NewEMA(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*e = ne

	return nil
}

// MarshalJSON converts EMA configuration data into JSON.
func (e EMA) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: e.length,
		Offset: e.offset,
	})
}

// namedMarshalJSON converts EMA configuration data with its
// name into JSON.
func (e EMA) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameEMA,
		Length: e.length,
		Offset: e.offset,
	})
}

// NameHMA returns HMA indicator name.
const NameHMA = "hma"

// HMA holds all the necessary information needed to calculate
// hull moving average.
// The zero value is not usable.
type HMA struct {
	// valid specifies whether HMA paremeters were validated.
	valid bool

	// wma specifies the base moving average.
	wma WMA
}

// NewHMA validates provided configuration options and
// creates new HMA indicator.
func NewHMA(w WMA) (HMA, error) {
	h := HMA{wma: w}

	if err := h.validate(); err != nil {
		return HMA{}, err
	}

	return h, nil
}

// Equal checks whether provided hma has exactly the same values as main
// hma.
func (h HMA) Equal(h1 HMA) bool {
	if h.valid != h1.valid {
		return false
	}

	return h.wma.Equal(h1.wma)
}

func (h HMA) equal(i Indicator) bool {
	h1, ok := i.(HMA)
	if ok {
		return h.Equal(h1)
	}

	return ok
}

// WMA returns wma configuration option.
func (h HMA) WMA() WMA {
	return h.wma
}

// Offset returns offset configuration option.
func (h HMA) Offset() int {
	return h.wma.offset
}

// validate checks whether HMA was configured properly or not.
func (h *HMA) validate() error {
	if err := h.wma.validate(); err != nil {
		return errors.New("invalid wma")
	}

	if h.wma.length < 2 {
		return ErrInvalidLength
	}

	h.valid = true

	return nil
}

// Calc calculates HMA from the provided data points slice.
// Calculation is based on formula provided by fidelity.
// https://www.fidelity.com/learning-center/trading-investing/technical-analysis/technical-indicator-guide/hull-moving-average.
// All credits are due to Alan Hull who developed HMA indicator.
func (h HMA) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !h.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, h.Count(), 0)
	if err != nil {
		return decimal.Zero, err
	}

	l := int(math.Sqrt(float64(h.wma.Count())))

	w1 := WMA{length: h.wma.Count() / 2, valid: true}
	w2 := h.wma
	w3 := WMA{length: l, valid: true}

	v := make([]decimal.Decimal, l)

	for i := 0; i < l; i++ {
		r1, _ := w1.Calc(dd[:len(dd)-l+i+1])

		r2, _ := w2.Calc(dd[:len(dd)-l+i+1])

		v[i] = r1.Mul(decimal.NewFromInt(2)).Sub(r2)
	}

	r, _ := w3.Calc(v)

	return r, nil
}

// Count determines the total amount of data points needed for HMA
// calculation.
func (h HMA) Count() int {
	return h.wma.Count()*2 - h.wma.offset - 1
}

// UnmarshalJSON parses JSON into HMA structure.
func (h *HMA) UnmarshalJSON(d []byte) error {
	var data struct {
		WMA struct {
			Length int `json:"length"`
			Offset int `json:"offset"`
		} `json:"wma"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	w, err := NewWMA(data.WMA.Length, data.WMA.Offset)
	if err != nil {
		return err
	}

	nh, err := NewHMA(w)
	if err != nil {
		return err
	}

	*h = nh

	return nil
}

// MarshalJSON converts HMA configuration data into JSON.
func (h HMA) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		WMA WMA `json:"wma"`
	}{
		WMA: h.wma,
	})
}

// namedMarshalJSON converts HMA configuration data with its
// name into JSON.
func (h HMA) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name String `json:"name"`
		WMA  WMA    `json:"wma"`
	}{
		Name: NameHMA,
		WMA:  h.wma,
	})
}

// NameCD returns CD indicator name.
const NameCD = "cd"

// CD holds all the necessary information needed to calculate
// difference between two source indicators.
// The zero value is not usable.
type CD struct {
	// valid specifies whether CD paremeters were validated.
	valid bool

	// source1 specifies which indicator to use as base
	// during calculation process.
	source1 Indicator

	// source2 specifies which indicator to use as counter
	// during calculation process.
	source2 Indicator

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewCD validates provided configuration options and
// creates new CD indicator.
func NewCD(source1, source2 Indicator, offset int) (CD, error) {
	cd := CD{source1: source1, source2: source2, offset: offset}

	if err := cd.validate(); err != nil {
		return CD{}, err
	}

	return cd, nil
}

// Equal checks whether provided cd has exactly the same values as main
// cd.
func (cd CD) Equal(cd1 CD) bool {
	if cd.valid != cd1.valid || cd.offset != cd1.offset {
		return false
	}

	return cd.source1.equal(cd1.source1) && cd.source2.equal(cd1.source2)
}

func (cd CD) equal(i Indicator) bool {
	cd1, ok := i.(CD)
	if ok {
		return cd.Equal(cd1)
	}

	return ok
}

// Sub1 returns source1 configuration option.
func (cd CD) Sub1() Indicator {
	return cd.source1
}

// Sub2 returns source2 configuration option.
func (cd CD) Sub2() Indicator {
	return cd.source2
}

// Offset returns offset configuration option.
func (cd CD) Offset() int {
	return cd.offset
}

// validate checks whether CD was configured properly or not.
func (cd *CD) validate() error {
	if cd.source1 == nil || cd.source2 == nil {
		return ErrInvalidSource
	}

	if cd.offset < 0 {
		return ErrInvalidOffset
	}

	cd.valid = true

	return nil
}

// Calc calculates CD from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/m/macd.asp.
// Formula has been improved upon so any indicators can be compared
// with each other.
func (cd CD) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !cd.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, cd.Count()-cd.offset, cd.offset)
	if err != nil {
		return decimal.Zero, err
	}

	r1, err := cd.source1.Calc(dd)
	if err != nil {
		return decimal.Zero, err
	}

	r2, err := cd.source2.Calc(dd)
	if err != nil {
		return decimal.Zero, err
	}

	r := r1.Sub(r2)

	return r, nil
}

// Count determines the total amount of data points needed for CD
// calculation.
func (cd CD) Count() int {
	c1 := cd.source1.Count()
	c2 := cd.source2.Count()

	if c1 > c2 {
		return c1 + cd.offset
	}

	return c2 + cd.offset
}

// UnmarshalJSON parses JSON into CD structure.
func (cd *CD) UnmarshalJSON(d []byte) error {
	var data struct {
		Source1 json.RawMessage `json:"source1"`
		Source2 json.RawMessage `json:"source2"`
		Offset  int             `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	src1, err := fromJSON(data.Source1)
	if err != nil {
		return err
	}

	src2, err := fromJSON(data.Source2)
	if err != nil {
		return err
	}

	nm, _ := NewCD(src1, src2, data.Offset)
	if err := nm.validate(); err != nil {
		return err
	}

	*cd = nm

	return nil
}

// MarshalJSON converts CD configuration data into JSON.
func (cd CD) MarshalJSON() ([]byte, error) {
	src1, err := cd.source1.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	src2, err := cd.source2.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Source1 json.RawMessage `json:"source1"`
		Source2 json.RawMessage `json:"source2"`
		Offset  int             `json:"offset"`
	}{
		Source1: src1,
		Source2: src2,
		Offset:  cd.offset,
	})
}

// namedMarshalJSON converts CD configuration data with its
// name into JSON.
func (cd CD) namedMarshalJSON() ([]byte, error) {
	src1, err := cd.source1.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	src2, err := cd.source2.namedMarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Name    String          `json:"name"`
		Source1 json.RawMessage `json:"source1"`
		Source2 json.RawMessage `json:"source2"`
		Offset  int             `json:"offset"`
	}{
		Name:    NameCD,
		Source1: src1,
		Source2: src2,
		Offset:  cd.offset,
	})
}

// NameROC returns ROC indicator name.
const NameROC = "roc"

// ROC holds all the necessary information needed to calculate rate
// of change.
// The zero value is not usable.
type ROC struct {
	// valid specifies whether ROC paremeters were validated.
	valid bool

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewROC validates provided configuration options and
// creates new ROC indicator.
func NewROC(length, offset int) (ROC, error) {
	r := ROC{length: length, offset: offset}

	if err := r.validate(); err != nil {
		return ROC{}, err
	}

	return r, nil
}

// Equal checks whether provided roc has exactly the same values as main
// roc.
func (r ROC) Equal(r1 ROC) bool {
	return r == r1
}

func (r ROC) equal(i Indicator) bool {
	r1, ok := i.(ROC)
	if ok {
		return r.Equal(r1)
	}

	return ok
}

// Length returns length configuration option.
func (r ROC) Length() int {
	return r.length
}

// Offset returns offset configuration option.
func (r ROC) Offset() int {
	return r.offset
}

// validate checks whether ROC was configured properly or not.
func (r *ROC) validate() error {
	if r.length < 1 {
		return ErrInvalidLength
	}

	if r.offset < 0 {
		return ErrInvalidOffset
	}

	r.valid = true

	return nil
}

// Calc calculates ROC from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/p/pricerateofchange.asp.
func (r ROC) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !r.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, r.Count()-r.offset, r.offset)
	if err != nil {
		return decimal.Zero, err
	}

	n := dd[len(dd)-1]
	l := dd[len(dd)-r.length]

	if l.Equal(decimal.Zero) {
		return decimal.Zero, nil
	}

	return n.Sub(l).Div(l).Mul(Hundred), nil
}

// Count determines the total amount of data points needed for ROC
// calculation.
func (r ROC) Count() int {
	return r.length + r.offset
}

// UnmarshalJSON parses JSON into ROC structure.
func (r *ROC) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	nr, err := NewROC(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*r = nr

	return nil
}

// MarshalJSON converts ROC configuration data into JSON.
func (r ROC) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: r.length,
		Offset: r.offset,
	})
}

// namedMarshalJSON converts ROC configuration data with its
// name into JSON.
func (r ROC) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameROC,
		Length: r.length,
		Offset: r.offset,
	})
}

// NameRSI returns RSI indicator name.
const NameRSI = "rsi"

// RSI holds all the necessary information needed to calculate relative
// strength index.
// The zero value is not usable.
type RSI struct {
	// valid specifies whether RSI paremeters were validated.
	valid bool

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewRSI validates provided configuration options and
// creates new RSI indicator.
func NewRSI(length, offset int) (RSI, error) {
	r := RSI{length: length, offset: offset}

	if err := r.validate(); err != nil {
		return RSI{}, err
	}

	return r, nil
}

// Equal checks whether provided rsi has exactly the same values as main
// rsi.
func (r RSI) Equal(r1 RSI) bool {
	return r == r1
}

func (r RSI) equal(i Indicator) bool {
	r1, ok := i.(RSI)
	if ok {
		return r.Equal(r1)
	}

	return ok
}

// Length returns length configuration option.
func (r RSI) Length() int {
	return r.length
}

// Offset returns offset configuration option.
func (r RSI) Offset() int {
	return r.offset
}

// validate checks whether RSI was configured properly or not.
func (r *RSI) validate() error {
	if r.length < 1 {
		return ErrInvalidLength
	}

	if r.offset < 0 {
		return ErrInvalidOffset
	}

	r.valid = true

	return nil
}

// Calc calculates RSI from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/r/rsi.asp.
// All credits are due to J. Welles Wilder Jr. who developed RSI indicator.
func (r RSI) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !r.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, r.Count()-r.offset, r.offset)
	if err != nil {
		return decimal.Zero, err
	}

	ag := decimal.Zero
	al := decimal.Zero
	length := decimal.NewFromInt(int64(r.length))

	for i := 1; i < len(dd); i++ {
		if dd[i].Sub(dd[i-1]).LessThan(decimal.Zero) {
			al = al.Add(dd[i].Sub(dd[i-1]).Abs())
		} else {
			ag = ag.Add(dd[i].Sub(dd[i-1]))
		}
	}

	if ag == decimal.Zero {
		return decimal.NewFromInt(0), nil
	}

	if al == decimal.Zero {
		return Hundred, nil
	}

	ag = ag.Div(length)

	al = al.Div(length)

	return Hundred.Sub(Hundred.Div(decimal.NewFromInt(1).Add(ag.Div(al)))), nil
}

// Count determines the total amount of data points needed for RSI
// calculation.
func (r RSI) Count() int {
	return r.length + r.offset
}

// UnmarshalJSON parses JSON into RSI structure.
func (r *RSI) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	nr, err := NewRSI(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*r = nr

	return nil
}

// MarshalJSON converts RSI configuration data into JSON.
func (r RSI) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: r.length,
		Offset: r.offset,
	})
}

// namedMarshalJSON converts RSI configuration data with its
// name into JSON.
func (r RSI) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameRSI,
		Length: r.length,
		Offset: r.offset,
	})
}

// NameSMA returns SMA indicator name.
const NameSMA = "sma"

// SMA holds all the necessary information needed to calculate simple
// moving average.
// The zero value is not usable.
type SMA struct {
	// valid specifies whether SMA paremeters were validated.
	valid bool

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewSMA validates provided configuration options and
// creates new SMA indicator.
func NewSMA(length, offset int) (SMA, error) {
	s := SMA{length: length, offset: offset}

	if err := s.validate(); err != nil {
		return SMA{}, err
	}

	return s, nil
}

// Equal checks whether provided sma has exactly the same values as main
// sma.
func (s SMA) Equal(s1 SMA) bool {
	return s == s1
}

func (s SMA) equal(i Indicator) bool {
	s1, ok := i.(SMA)
	if ok {
		return s.Equal(s1)
	}

	return ok
}

// Length returns length configuration option.
func (s SMA) Length() int {
	return s.length
}

// Offset returns offset configuration option.
func (s SMA) Offset() int {
	return s.offset
}

// validate checks whether SMA was configured properly or not.
func (s *SMA) validate() error {
	if s.length < 1 {
		return ErrInvalidLength
	}

	if s.offset < 0 {
		return ErrInvalidOffset
	}

	s.valid = true

	return nil
}

// Calc calculates SMA from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/s/sma.asp.
func (s SMA) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !s.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, s.Count()-s.offset, s.offset)
	if err != nil {
		return decimal.Zero, err
	}

	r := decimal.Zero

	for i := 0; i < len(dd); i++ {
		r = r.Add(dd[i])
	}

	return r.Div(decimal.NewFromInt(int64(s.length))), nil
}

// Count determines the total amount of data points needed for SMA
// calculation.
func (s SMA) Count() int {
	return s.length + s.offset
}

// UnmarshalJSON parses JSON into SMA structure.
func (s *SMA) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	ns, err := NewSMA(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*s = ns

	return nil
}

// MarshalJSON converts SMA configuration data into JSON.
func (s SMA) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: s.length,
		Offset: s.offset,
	})
}

// namedMarshalJSON converts SMA configuration data with its
// name into JSON.
func (s SMA) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameSMA,
		Length: s.length,
		Offset: s.offset,
	})
}

// NameSRSI returns SRSI indicator name.
const NameSRSI = "srsi"

// SRSI holds all the necessary information needed to calculate stoch
// relative strength index.
// The zero value is not usable.
type SRSI struct {
	// valid specifies whether SRSI paremeters were validated.
	valid bool

	// rsi specifies the base relative strength index.
	rsi RSI
}

// NewSRSI validates provided configuration options and
// creates new SRSI indicator.
func NewSRSI(r RSI) (SRSI, error) {
	s := SRSI{rsi: r}

	if err := s.validate(); err != nil {
		return SRSI{}, err
	}

	return s, nil
}

// Equal checks whether provided srsi has exactly the same values as main
// srsi.
func (s SRSI) Equal(s1 SRSI) bool {
	if s.valid != s1.valid {
		return false
	}

	return s.rsi.Equal(s1.rsi)
}

func (s SRSI) equal(i Indicator) bool {
	s1, ok := i.(SRSI)
	if ok {
		return s.Equal(s1)
	}

	return ok
}

// RSI returns rsi configuration option.
func (s SRSI) RSI() RSI {
	return s.rsi
}

// Offset returns offset configuration option.
func (s SRSI) Offset() int {
	return s.rsi.offset
}

// validate checks whether SRSI was configured properly or not.
func (s *SRSI) validate() error {
	if err := s.rsi.validate(); err != nil {
		return err
	}

	s.valid = true

	return nil
}

// Calc calculates SRSI from the provided data slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/s/stochrsi.asp.
func (s SRSI) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !s.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	v, err := calcMultiple(s.rsi, dd, s.rsi.length)
	if err != nil {
		return decimal.Zero, err
	}

	c := v[0]
	h := v[0]
	l := v[0]

	for i := 1; i < len(v); i++ {
		if h.LessThan(v[i]) {
			h = v[i]
		}

		if l.GreaterThan(v[i]) {
			l = v[i]
		}
	}

	denom := h.Sub(l)
	if denom.Equal(decimal.Zero) {
		return decimal.Zero, nil
	}

	return c.Sub(l).Div(denom), nil
}

// Count determines the total amount of data needed for SRSI
// calculation.
func (s SRSI) Count() int {
	return s.rsi.length*2 + s.rsi.offset - 1
}

// UnmarshalJSON parses JSON into SRSI structure.
func (s *SRSI) UnmarshalJSON(d []byte) error {
	var data struct {
		RSI struct {
			Length int `json:"length"`
			Offset int `json:"offset"`
		} `json:"rsi"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	r, err := NewRSI(data.RSI.Length, data.RSI.Offset)
	if err != nil {
		return err
	}

	ns, _ := NewSRSI(r)

	*s = ns

	return nil
}

// MarshalJSON converts SRSI configuration data into JSON.
func (s SRSI) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		RSI RSI `json:"rsi"`
	}{
		RSI: s.rsi,
	})
}

// namedMarshalJSON converts SRSI configuration data with its
// name into JSON.
func (s SRSI) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name String `json:"name"`
		RSI  RSI    `json:"rsi"`
	}{
		Name: NameSRSI,
		RSI:  s.rsi,
	})
}

// NameStoch returns Stoch  indicator name.
const NameStoch = "stoch"

// Stoch holds all the necessary information needed to calculate stochastic
// oscillator.
// The zero value is not usable.
type Stoch struct {
	// valid specifies whether Stoch paremeters were validated.
	valid bool

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewStoch validates provided configuration options and
// creates new Stoch indicator.
func NewStoch(length, offset int) (Stoch, error) {
	s := Stoch{length: length, offset: offset}

	if err := s.validate(); err != nil {
		return Stoch{}, err
	}

	return s, nil
}

// Equal checks whether provided stoch has exactly the same values as main
// stoch.
func (s Stoch) Equal(s1 Stoch) bool {
	return s == s1
}

func (s Stoch) equal(i Indicator) bool {
	s1, ok := i.(Stoch)
	if ok {
		return s.Equal(s1)
	}

	return ok
}

// Length returns length configuration option.
func (s Stoch) Length() int {
	return s.length
}

// Offset returns offset configuration option.
func (s Stoch) Offset() int {
	return s.offset
}

// validate checks whether Stoch was configured properly or not.
func (s *Stoch) validate() error {
	if s.length < 1 {
		return ErrInvalidLength
	}

	if s.offset < 0 {
		return ErrInvalidOffset
	}

	s.valid = true

	return nil
}

// Calc calculates Stoch from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/terms/s/stochasticoscillator.asp.
func (s Stoch) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !s.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, s.Count()-s.offset, s.offset)
	if err != nil {
		return decimal.Zero, err
	}

	l := dd[0]
	h := dd[0]

	for i := 0; i < len(dd); i++ {
		if dd[i].LessThan(l) {
			l = dd[i]
		}

		if dd[i].GreaterThan(h) {
			h = dd[i]
		}
	}

	denom := h.Sub(l)
	if denom.Equal(decimal.Zero) {
		return decimal.Zero, nil
	}

	return dd[len(dd)-1].Sub(l).Div(denom).Mul(Hundred), nil
}

// Count determines the total amount of data points needed for Stoch
// calculation.
func (s Stoch) Count() int {
	return s.length + s.offset
}

// UnmarshalJSON parses JSON into Stoch structure.
func (s *Stoch) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	ns, err := NewStoch(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*s = ns

	return nil
}

// MarshalJSON converts Stoch configuration data into JSON.
func (s Stoch) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: s.length,
		Offset: s.offset,
	})
}

// namedMarshalJSON converts Stoch configuration data with its
// name into JSON.
func (s Stoch) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameStoch,
		Length: s.length,
		Offset: s.offset,
	})
}

// NameWMA returns WMA  indicator name.
const NameWMA = "wma"

// WMA holds all the necessary information needed to calculate weighted
// moving average.
// The zero value is not usable.
type WMA struct {
	// valid specifies whether WMA paremeters were validated.
	valid bool

	// length specifies how many data points should be used
	// during the calculations.
	length int

	// offset specifies how many data points should be skipped from the end
	// during the calculations.
	offset int
}

// NewWMA validates provided configuration options and
// creates new WMA indicator.
func NewWMA(length, offset int) (WMA, error) {
	w := WMA{length: length, offset: offset}

	if err := w.validate(); err != nil {
		return WMA{}, err
	}

	return w, nil
}

// Equal checks whether provided wma has exactly the same values as main
// wma.
func (w WMA) Equal(w1 WMA) bool {
	return w == w1
}

func (w WMA) equal(i Indicator) bool {
	w1, ok := i.(WMA)
	if ok {
		return w.Equal(w1)
	}

	return ok
}

// Length returns length configuration option.
func (w WMA) Length() int {
	return w.length
}

// Offset returns offset configuration option.
func (w WMA) Offset() int {
	return w.offset
}

// validate checks whether WMA was configured properly or not.
func (w *WMA) validate() error {
	if w.length < 1 {
		return ErrInvalidLength
	}

	if w.offset < 0 {
		return ErrInvalidOffset
	}

	w.valid = true

	return nil
}

// Calc calculates WMA from the provided data points slice.
// Calculation is based on formula provided by investopedia.
// https://www.investopedia.com/articles/technical/060401.asp.
func (w WMA) Calc(dd []decimal.Decimal) (decimal.Decimal, error) {
	if !w.valid {
		return decimal.Zero, ErrInvalidIndicator
	}

	dd, err := resize(dd, w.Count()-w.offset, w.offset)
	if err != nil {
		return decimal.Zero, err
	}

	r := decimal.Zero

	wi := decimal.NewFromInt(int64(w.length * (w.length + 1))).Div(decimal.NewFromInt(2))

	for i := 0; i < len(dd); i++ {
		r = r.Add(dd[i].Mul(decimal.NewFromInt(int64(i + 1)).Div(wi)))
	}

	return r, nil
}

// Count determines the total amount of data points needed for WMA
// calculation.
func (w WMA) Count() int {
	return w.length + w.offset
}

// UnmarshalJSON parses JSON into WMA structure.
func (w *WMA) UnmarshalJSON(d []byte) error {
	var data struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}

	if err := json.Unmarshal(d, &data); err != nil {
		return err
	}

	nw, err := NewWMA(data.Length, data.Offset)
	if err != nil {
		return err
	}

	*w = nw

	return nil
}

// MarshalJSON converts WMA configuration data into JSON.
func (w WMA) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Length int `json:"length"`
		Offset int `json:"offset"`
	}{
		Length: w.length,
		Offset: w.offset,
	})
}

// namedMarshalJSON converts WMA configuration data with its
// name into JSON.
func (w WMA) namedMarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name   String `json:"name"`
		Length int    `json:"length"`
		Offset int    `json:"offset"`
	}{
		Name:   NameWMA,
		Length: w.length,
		Offset: w.offset,
	})
}
