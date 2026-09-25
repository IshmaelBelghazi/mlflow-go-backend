package query_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mlflow/mlflow-go-backend/pkg/tracking/service/query"
)

func TestValidQueries(t *testing.T) {
	t.Parallel()

	samples := []string{
		"metrics.foobar = 40",
		"metrics.foobar = 40 AND run_name = \"bouncy-boar-498\"",
		"tags.\"mlflow.source.name\" = \"scratch.py\"",
		"tags.`mlflow.runName` IS NULL",
		"metrics.accuracy > 0.9",
		"params.\"random_state\" = \"8888\"",
		"params.`random_state` = \"8888\"",
		"params.`random_state` IS NOT NULL",
		"params.solver ILIKE \"L%\"",
		"params.solver LIKE \"l%\"",
		"datasets.digest IN ('77a19fc0')",
		"attributes.run_id IN ('meh')",
	}

	for _, sample := range samples {
		currentSample := sample
		t.Run(currentSample, func(t *testing.T) {
			t.Parallel()

			_, err := query.ParseFilter(currentSample)
			if err != nil {
				t.Errorf("unexpected parse error: %v", err)
			}
		})
	}
}

type invalidSample struct {
	input         string
	expectedError string
}

//nolint:funlen
func TestInvalidQueries(t *testing.T) {
	t.Parallel()

	samples := []invalidSample{
		{
			input:         "yow.foobar = 40",
			expectedError: "invalid identifier",
		},
		{
			input: "attributes.foobar = 40",
			expectedError: "Invalid attribute key '{foobar}' specified. " +
				"Valid keys are '[run_id run_name user_id status start_time end_time artifact_uri]'",
		},
		{
			input: "datasets.foobar = 40",
			expectedError: "Invalid dataset key '{foobar}' specified. " +
				"Valid keys are '[run_id run_name user_id status start_time end_time artifact_uri]'",
		},
		{
			input:         "metric.yow = 'z'",
			expectedError: "expected numeric value type for metric.",
		},
		{
			input:         "parameter.tag = 2",
			expectedError: "expected a quoted string value",
		},
		{
			input:         "attributes.start_time = 'now'",
			expectedError: "expected numeric value type for numeric attribute",
		},
		{
			input:         "attributes.run_name IN ('foo','bar')",
			expectedError: "only the 'run_id' attribute supports comparison with a list",
		},
		{
			input:         "attributes.status IS NULL",
			expectedError: "IS NULL / IS NOT NULL is only supported for tags and params",
		},
		{
			input:         "datasets.name = 40",
			expectedError: "expected datasets.name to be either a string or list of strings",
		},
		{
			input:         "datasets.digest = 50",
			expectedError: "expected datasets.digest to be either a string or list of strings",
		},
		{
			input:         "datasets.context = 60",
			expectedError: "expected datasets.context to be either a string or list of strings",
		},
	}

	for _, sample := range samples {
		currentSample := sample
		t.Run(currentSample.input, func(t *testing.T) {
			t.Parallel()

			_, err := query.ParseFilter(currentSample.input)
			if err == nil {
				t.Errorf("expected parse error but got nil")
			}

			if !strings.Contains(err.Error(), currentSample.expectedError) {
				t.Errorf(
					"expected error to contain %q, got %q",
					currentSample.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestDottedIdentifiersMatchQuotedKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		namespace string
		key       string
		tail      string
	}{
		{"tags", "life.worker", " = 'Trainer'"},
		{"tag", "life.worker.role", " != 'Evaluator'"},
		{"tags", "mlflow.runName", " LIKE 'train%'"},
		{"params", "model.optim.lr", " = '0.001'"},
		{"metrics", "valid.loss", " > -0.5 AND metrics.train.loss < 0.25"},
		{"tags", "life.worker", " IS NULL"},
		{"params", "model.optim.lr", " IS NOT NULL"},
		{"tags", "life.AND", " = 'Trainer' AND attributes.status = 'FINISHED'"},
		{"tags", "life.NULL", " = 'Trainer'"},
	}
	for _, sample := range cases {
		t.Run(sample.namespace+"."+sample.key+sample.tail, func(t *testing.T) {
			plain, err := query.ParseFilter(sample.namespace + "." + sample.key + sample.tail)
			if err != nil {
				t.Fatal(err)
			}
			for _, quote := range []string{"`", "\"", "'"} {
				quoted, err := query.ParseFilter(sample.namespace + "." + quote + sample.key + quote + sample.tail)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(plain, quoted) || plain[0].Key != sample.key {
					t.Fatalf("plain and quoted keys differ: %#v versus %#v", plain, quoted)
				}
			}
		})
	}
}

func TestMalformedDottedIdentifiersReturnErrors(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"tags.life.worker", "tags.life.worker =", "tags.life.worker = 'Trainer' AND",
		"tags.life worker = 'Trainer'", "tags.life.123 = 'Trainer'",
		"attributes.life.worker = 'Trainer'", "datasets.life.worker = 'Trainer'",
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := query.ParseFilter(input); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
