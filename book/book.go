package book

import (
	"bytes"
	"fmt"

	"github.com/nikandfor/skiplist"
)

type (
	ID uint64

	OrderType bool

	Rat int64 // *1e9

	Price  = Rat
	Amount = Rat

	Order struct {
		ID
		Type OrderType
		Price
		Amount

		Filled Amount
		Money  Amount
	}

	Book struct {
		bids *skiplist.List
		asks *skiplist.List
		last Price
	}

	DepthPoint struct {
		Price
		Amount
	}

	Depth struct {
		Asks []DepthPoint
		Bids []DepthPoint
	}
)

const (
	Ask OrderType = iota != 0
	Bid
)

const (
	Div  = 1e9
	Zero = Rat(0)
	One  = 1 * Div
)

func New() *Book {
	b := &Book{
		asks: skiplist.NewRepeated(func(a, b interface{}) bool {
			return a.(*Order).Price > b.(*Order).Price
		}),
		bids: skiplist.NewRepeated(func(a, b interface{}) bool {
			return a.(*Order).Price < b.(*Order).Price
		}),
	}
	b.asks.SetAutoReuse(false)
	b.bids.SetAutoReuse(false)
	return b
}

func (b *Book) Trade(o *Order) []*Order {
	updated := []*Order{o}

	var fill *skiplist.List
	var addto *skiplist.List

	if o.Type == Ask {
		fill = b.bids
		addto = b.asks
	} else {
		fill = b.asks
		addto = b.bids
	}

	for e := fill.First(); e != nil && o.Filled < o.Amount; e = e.Next() {
		cur := e.Value().(*Order)
		if o.Type == Ask {
			if cur.Price > o.Price {
				break
			}
		} else {
			if cur.Price < o.Price {
				break
			}
		}

		var dialAmount Amount
		if o.Amount-o.Filled > cur.Amount-cur.Filled {
			dialAmount = cur.Amount - cur.Filled
			fill.Del(cur)
		} else {
			dialAmount = o.Amount - o.Filled
		}
		money := dialAmount.Mul(cur.Price)
		o.Filled += dialAmount
		cur.Filled += dialAmount
		o.Money += money
		cur.Money += money

		updated = append(updated, cur)

		b.last = cur.Price
	}

	if o.Filled < o.Amount {
		addto.Put(o)
	}

	if len(updated) != 1 {
		return updated
	}

	return nil
}

func (b *Book) Cancel(o *Order) *Order {
	var l *skiplist.List
	if o.Type == Ask {
		l = b.asks
	} else {
		l = b.bids
	}
	el := l.DelIf(o, func(el *skiplist.El) bool {
		cur := el.Value().(*Order)
		return cur.ID == o.ID
	})
	if el == nil {
		return nil
	}
	res := el.Value()
	skiplist.Reuse(el)
	return res.(*Order)
}

func (b *Book) MiddlePrice() Price {
	var hi, lo Price
	if e := b.bids.First(); e != nil {
		cur := e.Value().(*Order)
		hi = cur.Price
	}

	if e := b.asks.First(); e != nil {
		cur := e.Value().(*Order)
		lo = cur.Price
	} else {
		lo = hi
	}

	if hi == 0 {
		hi = lo
	}

	return hi/2 + lo/2
}

func (b *Book) LastPrice() Price {
	return b.last
}

func (b *Book) Depth(n int) Depth {
	f := func(l *skiplist.List) []DepthPoint {
		res := make([]DepthPoint, n)
		i := 0
		var add DepthPoint
		for e := l.First(); e != nil && i != n; e = e.Next() {
			cur := e.Value().(*Order)
			if cur.Price != add.Price {
				if add.Price != 0 {
					res[i] = add
					i++
				}
				add = DepthPoint{Price: cur.Price}
			}
			add.Amount += (cur.Amount - cur.Filled)
		}
		if i != n && add.Amount != 0 {
			res[i] = add
			i++
		}
		if i < len(res) {
			res = res[:i]
		}
		return res
	}

	return Depth{Asks: f(b.asks), Bids: f(b.bids)}
}

func (b *Book) Dump() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "asks (%d):\n", b.asks.Len())
	for e := b.asks.First(); e != nil; e = e.Next() {
		cur := e.Value().(*Order)
		fmt.Fprintf(&buf, "  %v\n", cur)
	}
	fmt.Fprintf(&buf, "bids (%d):\n", b.bids.Len())
	for e := b.bids.First(); e != nil; e = e.Next() {
		cur := e.Value().(*Order)
		fmt.Fprintf(&buf, "  %v\n", cur)
	}
	return buf.String()
}

func (d Depth) Dump() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "asks (%d):\n", len(d.Asks))
	for _, p := range d.Asks {
		fmt.Fprintf(&buf, "  %v %v\n", p.Price, p.Amount)
	}
	fmt.Fprintf(&buf, "bids (%d):\n", len(d.Bids))
	for _, p := range d.Bids {
		fmt.Fprintf(&buf, "  %v %v\n", p.Price, p.Amount)
	}
	return buf.String()
}

func (o *Order) String() string {
	return fmt.Sprintf("%v %v pr %v amount %v (filled %v for %v)", o.Type, o.ID, o.Price, o.Amount, o.Filled, o.Money)
}

func RatFromInt(v, d int64) Rat {
	return Rat(v * Div / d)
}
func RatFromFloat(f float64) Rat {
	return Rat(f * Div)
}

func (r Rat) String() string {
	return fmt.Sprintf("%15.5f", float64(r)/Div)
}

func (r Rat) Mul(b Rat) Rat {
	r /= 1e5
	b /= 1e4
	return r * b
}

func (d ID) String() string {
	return fmt.Sprintf("%16x", uint64(d))
}

func (t OrderType) String() string {
	if t == Ask {
		return "ask"
	} else {
		return "bid"
	}
}
