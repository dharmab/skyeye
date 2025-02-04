package controller

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	alphaCheckCounter metric.Int64Counter
	bogeyDopeCounter  metric.Int64Counter
	checkInCounter    metric.Int64Counter
	declareCounter    metric.Int64Counter
	pictureCounter    metric.Int64Counter
	radioCheckCounter metric.Int64Counter
	shoppingCounter   metric.Int64Counter
	snaplockCounter   metric.Int64Counter
	spikedCounter     metric.Int64Counter
	tripwireCounter   metric.Int64Counter
	unableCounter     metric.Int64Counter
)

func init() {
	meter := otel.Meter("")
	initRequestCounter(meter, &alphaCheckCounter, "controller.alphacheck.counter", "Number of ALPHA CHECK requests handled by the controller")
	initRequestCounter(meter, &bogeyDopeCounter, "controller.bogeydope.counter", "Number of BOGEY DOPE requests handled by the controller")
	initRequestCounter(meter, &checkInCounter, "controller.checkin.counter", "Number of ambiguous check-in requests handled by the controller")
	initRequestCounter(meter, &declareCounter, "controller.declare.counter", "Number of DECLARE requests handled by the controller")
	initRequestCounter(meter, &pictureCounter, "controller.picture.counter", "Number of PICTURE requests handled by the controller")
	initRequestCounter(meter, &radioCheckCounter, "controller.radiocheck.counter", "Number of RADIO CHECK requests handled by the controller")
	initRequestCounter(meter, &shoppingCounter, "controller.shopping.counter", "Number of SHOPPING requests handled by the controller")
	initRequestCounter(meter, &snaplockCounter, "controller.snaplock.counter", "Number of SNAPLOCK requests handled by the controller")
	initRequestCounter(meter, &spikedCounter, "controller.spiked.counter", "Number of SPIKED requests handled by the controller")
	initRequestCounter(meter, &tripwireCounter, "controller.tripwire.counter", "Number of TRIPWIRE requests handled by the controller")
	initRequestCounter(meter, &unableCounter, "controller.unable.counter", "Number of UNABLE requests handled by the controller")
}

func initRequestCounter(meter metric.Meter, counter *metric.Int64Counter, name, description string) {
	requestCounter, err := meter.Int64Counter(
		name,
		metric.WithDescription(description),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}
	*counter = requestCounter
}
