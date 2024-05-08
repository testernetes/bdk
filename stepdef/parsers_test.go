package stepdef_test

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	tString = reflect.TypeOf("")
	tInt    = reflect.TypeOf(0)
	tBool   = reflect.TypeOf(false)
	tFloat  = reflect.TypeOf(0.1)

	tMatcher      = reflect.TypeOf((*types.GomegaMatcher)(nil)).Elem()
	tAssert       = reflect.TypeOf((stepdef.Assert)(nil))
	tDuration     = reflect.TypeOf(time.Duration(0))
	tUnstructured = reflect.TypeOf((*unstructured.Unstructured)(nil))
	tPod          = reflect.TypeOf((*corev1.Pod)(nil))

	tClientDryRun = reflect.TypeOf(client.DryRunAll)
)

var u = &unstructured.Unstructured{}
var pod = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}}

var _ = Describe("StringParser", Ordered, func() {
	var (
		ctxWithNoStore = context.Background()
		ctxWithStore   = store.NewStoreFor(context.Background())
	)

	BeforeAll(func() {
		u.SetName("my-pod")
		store.Save(ctxWithStore, "my-ref", u)
	})

	DescribeTable("Parsing a string into a given type",
		func(ctx context.Context, input string, t reflect.Type, expected any) {
			value, err := stepdef.StringParsers.Parse(ctx, input, t)
			Expect(err).ShouldNot(HaveOccurred())

			switch reflect.TypeOf(expected).Kind() {
			case reflect.Func:
				Expect(value.Pointer()).Should(Equal(reflect.ValueOf(expected).Pointer()))
			default:
				Expect(value.Interface()).Should(Equal(expected))
			}
		},
		Entry("Text Parser", ctxWithNoStore, "word", tString, "word"),
		Entry("Number Parser", ctxWithNoStore, "47", tInt, 47),
		Entry("Number Parser", ctxWithNoStore, "4.7", tFloat, 4.7),

		Entry("Boolean", ctxWithNoStore, "true", tBool, true),
		Entry("Boolean", ctxWithNoStore, "false", tBool, false),
		Entry("Boolean", ctxWithNoStore, "", tBool, false),

		Entry("Assert", ctxWithNoStore, "within", tAssert, stepdef.Eventually),
		Entry("Assert", ctxWithNoStore, "for at least", tAssert, stepdef.Consistently),

		Entry("Duration", ctxWithNoStore, "10s", tDuration, 10*time.Second),

		Entry("Unstructured", ctxWithStore, "my-ref", tUnstructured, u),
		Entry("Pod", ctxWithStore, "my-ref", tPod, pod),

		Entry("Matcher", ctxWithNoStore, "equal bar", tMatcher, BeEquivalentTo("bar")),
		Entry("Matcher", ctxWithNoStore, "be empty", tMatcher, BeEmpty()),
		Entry("Matcher", ctxWithNoStore, "have length 5", tMatcher, HaveLen(5)),
		Entry("Matcher", ctxWithNoStore, "say helloworld", tMatcher, gbytes.Say("helloworld")),

		Entry("ClientOption", ctxWithNoStore, "true", tClientDryRun, client.DryRunAll),
	)
})
