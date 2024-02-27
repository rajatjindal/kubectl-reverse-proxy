package dashboard

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

//go:embed metrics.prom
var rawd []byte

func TestOne(t *testing.T) {
	defer t.Fail()

	decoder := expfmt.NewDecoder(bytes.NewReader(rawd), expfmt.NewFormat(expfmt.TypeTextPlain))
	dec := &expfmt.SampleDecoder{
		Dec: decoder,
		Opts: &expfmt.DecodeOptions{
			Timestamp: model.Now(),
		},
	}

	var all model.Vector
	for {
		var smpls model.Vector
		err := dec.Decode(&smpls)
		if err != nil && errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, smpls...)
	}

	for _, v := range all {
		fmt.Println(v.String())

		if v.Histogram != nil {
			fmt.Println("inside histo")
			fmt.Println(v.Histogram.Buckets)
		}
	}
}

func TestTwo(t *testing.T) {
	defer t.Fail()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(bytes.NewReader(rawd))
	require.Nil(t, err)

	// opts := &expfmt.DecodeOptions{
	// 	Timestamp: model.Now(),
	// }

	mf2 := mf["caddy_reverse_proxy_upstreams_healthy"]
	fmt.Println(len(mf2.Metric))

	for key, f := range mf {
		fmt.Println("key is ", key)
		if f.GetType() == 4 {
			for _, m := range f.Metric {
				fmt.Println(*f.Name)
				fmt.Println(m.Histogram.String())
				for _, b := range m.Histogram.Bucket {
					fmt.Println(b.GetUpperBound(), " = ", b.GetCumulativeCount())
				}
			}
		}

		if f.GetType() == 0 {
			for _, m := range f.Metric {
				fmt.Println(*f.Name, m.Counter.String())
			}
		}
	}
}
