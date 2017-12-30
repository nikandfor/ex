package book

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rff = RatFromFloat
	rd  = RatFromInt
)

func TestDepth(t *testing.T) {
	b := New()

	b.Trade(&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100)})
	b.Trade(&Order{ID: 2, Type: Bid, Price: rff(20), Amount: rff(101)})
	b.Trade(&Order{ID: 3, Type: Bid, Price: rff(30), Amount: rff(102)})

	b.Trade(&Order{ID: 4, Type: Ask, Price: rff(9), Amount: rff(99)})
	b.Trade(&Order{ID: 5, Type: Ask, Price: rff(8), Amount: rff(98)})
	b.Trade(&Order{ID: 6, Type: Ask, Price: rff(7), Amount: rff(97)})

	d := b.Depth(10)
	assert.Equal(t, Depth{
		Bids: []DepthPoint{
			{Price: rff(10), Amount: rff(100)},
			{Price: rff(20), Amount: rff(101)},
			{Price: rff(30), Amount: rff(102)},
		},
		Asks: []DepthPoint{
			{Price: rff(9), Amount: rff(99)},
			{Price: rff(8), Amount: rff(98)},
			{Price: rff(7), Amount: rff(97)},
		},
	}, d)
}

func TestMiddlePrice(t *testing.T) {
	b := New()

	assert.Equal(t, rff(0), b.MiddlePrice())

	b.Trade(&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100)})

	assert.Equal(t, rff(10), b.MiddlePrice())

	b.Trade(&Order{ID: 4, Type: Ask, Price: rff(9), Amount: rff(50)})

	assert.Equal(t, rd(19, 2), b.MiddlePrice())
}

func TestLastPrice(t *testing.T) {
	b := New()

	assert.Equal(t, rff(0), b.LastPrice())

	b.Trade(&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100)})

	assert.Equal(t, rff(0), b.LastPrice())

	b.Trade(&Order{ID: 1, Type: Ask, Price: rff(20), Amount: rff(100)})

	assert.Equal(t, rff(10), b.LastPrice())
}

func TestAskBid(t *testing.T) {
	b := New()

	price := RatFromFloat(1)

	var up []*Order

	up = b.Trade(&Order{ID: 1, Type: Ask, Price: price, Amount: RatFromFloat(1)})
	assert.Equal(t, []*Order(nil), up)

	t.Logf("\n%v", b.Depth(2).Dump())
	t.Logf("\n%v", b.Dump())

	up = b.Trade(&Order{ID: 2, Type: Bid, Price: price, Amount: RatFromFloat(1)})
	assert.Equal(t, []*Order{
		{ID: 2, Type: Bid, Price: price, Amount: RatFromFloat(1), Filled: RatFromFloat(1), Money: RatFromFloat(1)},
		{ID: 1, Type: Ask, Price: price, Amount: RatFromFloat(1), Filled: RatFromFloat(1), Money: RatFromFloat(1)},
	}, up)

	t.Logf("\n%v", b.Depth(2).Dump())
	t.Logf("\n%v", b.Dump())
}

func TestBidAsk(t *testing.T) {
	b := New()

	price := RatFromFloat(1)

	var up []*Order

	up = b.Trade(&Order{ID: 1, Type: Bid, Price: price, Amount: RatFromFloat(1)})
	assert.Equal(t, []*Order(nil), up)

	t.Logf("\n%v", b.Depth(2).Dump())
	t.Logf("\n%v", b.Dump())

	up = b.Trade(&Order{ID: 2, Type: Ask, Price: price, Amount: RatFromFloat(1)})
	assert.Equal(t, []*Order{
		{ID: 2, Type: Ask, Price: price, Amount: RatFromFloat(1), Filled: RatFromFloat(1), Money: RatFromFloat(1)},
		{ID: 1, Type: Bid, Price: price, Amount: RatFromFloat(1), Filled: RatFromFloat(1), Money: RatFromFloat(1)},
	}, up)

	t.Logf("\n%v", b.Depth(2).Dump())
	t.Logf("\n%v", b.Dump())
}

func TestBidAskChain(t *testing.T) {
	b := New()

	b.Trade(&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100)})
	b.Trade(&Order{ID: 2, Type: Bid, Price: rff(15), Amount: rff(100)})
	b.Trade(&Order{ID: 3, Type: Bid, Price: rff(30), Amount: rff(100)})

	b.Trade(&Order{ID: 4, Type: Ask, Price: rff(5), Amount: rff(100)})

	t.Logf("\n%v", b.Dump())

	d := b.Depth(10)
	assert.Equal(t, Depth{
		Asks: []DepthPoint{
			{Price: rff(5), Amount: rff(100)},
		},
		Bids: []DepthPoint{
			{Price: rff(10), Amount: rff(100)},
			{Price: rff(15), Amount: rff(100)},
			{Price: rff(30), Amount: rff(100)},
		},
	}, d)

	tr1 := b.Trade(&Order{ID: 5, Type: Ask, Price: rff(20), Amount: rff(300)})
	assert.Equal(t, []*Order{
		&Order{ID: 5, Type: Ask, Price: rff(20), Amount: rff(300), Filled: rff(200), Money: rff(10*100 + 15*100)},
		&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100), Filled: rff(100), Money: rff(10 * 100)},
		&Order{ID: 2, Type: Bid, Price: rff(15), Amount: rff(100), Filled: rff(100), Money: rff(15 * 100)},
	}, tr1)

	t.Logf("\n%v", b.Dump())

	d = b.Depth(10)
	assert.Equal(t, Depth{
		Bids: []DepthPoint{
			{Price: rff(30), Amount: rff(100)},
		},
		Asks: []DepthPoint{
			{Price: rff(20), Amount: rff(100)},
			{Price: rff(5), Amount: rff(100)},
		},
	}, d)

	tr2 := b.Trade(&Order{ID: 6, Type: Bid, Price: rff(10), Amount: rff(200)})
	assert.Equal(t, []*Order{
		&Order{ID: 6, Type: Bid, Price: rff(10), Amount: rff(200), Filled: rff(100), Money: rff(20 * 100)},
		&Order{ID: 5, Type: Ask, Price: rff(20), Amount: rff(300), Filled: rff(300), Money: rff(10*100 + 15*100 + 20*100)},
	}, tr2)

	t.Logf("\n%v", b.Dump())

	d = b.Depth(10)
	assert.Equal(t, Depth{
		Bids: []DepthPoint{
			{Price: rff(10), Amount: rff(100)},
			{Price: rff(30), Amount: rff(100)},
		},
		Asks: []DepthPoint{
			{Price: rff(5), Amount: rff(100)},
		},
	}, d)
}

func TestCancel(t *testing.T) {
	b := New()

	b.Trade(&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100)})
	b.Trade(&Order{ID: 2, Type: Bid, Price: rff(10), Amount: rff(100)})
	b.Trade(&Order{ID: 3, Type: Bid, Price: rff(10), Amount: rff(100)})

	b.Cancel(&Order{ID: 2, Type: Bid, Price: rff(10)})

	tr := b.Trade(&Order{ID: 5, Type: Ask, Price: rff(10), Amount: rff(300)})
	assert.Equal(t, []*Order{
		&Order{ID: 5, Type: Ask, Price: rff(10), Amount: rff(300), Filled: rff(200), Money: rff(2 * 10 * 100)},
		&Order{ID: 1, Type: Bid, Price: rff(10), Amount: rff(100), Filled: rff(100), Money: rff(10 * 100)},
		&Order{ID: 3, Type: Bid, Price: rff(10), Amount: rff(100), Filled: rff(100), Money: rff(10 * 100)},
	}, tr)

	cl := b.Cancel(&Order{ID: 5, Type: Ask, Price: rff(10)})
	assert.Equal(t, &Order{ID: 5, Type: Ask, Price: rff(10), Amount: rff(300), Filled: rff(200), Money: rff(2 * 10 * 100)},
		cl)

	cl = b.Cancel(&Order{ID: 5, Type: Ask, Price: rff(10)})
	assert.Equal(t, (*Order)(nil), cl)
}

func BenchmarkTrade(b *testing.B) {
	e := New()
	price := rff(10)
	var vol Amount

	for i := 0; i < b.N; i++ {
		tp := Ask
		if i%2 == 0 {
			tp = Bid
		}

		diff := rand.Int63n(1000)
		if i>>1%2 == 0 {
			price = price.Mul(One - rd(diff, 100000))
		} else {
			diff += 100
			price = price.Mul(One + rd(diff, 100000))
		}

		am := rand.Int63n(1000)

		o := &Order{ID: ID(i + 1), Type: tp, Price: price, Amount: rd(am, 1)}

		tr := e.Trade(o)

		sum := Zero
		if len(tr) != 0 {
			sum = tr[0].Filled
			vol += sum
		}
		//b.Logf("vol %10v| %3d orders up by %v", vol, len(tr), o)

		price = e.MiddlePrice()
	}

	b.Logf("price: %v, volume: %v", price, vol)
}
