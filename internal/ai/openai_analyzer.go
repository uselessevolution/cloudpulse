package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type OpenAIAnalyzer struct {
	client openai.Client
	model  string
}

func NewOpenAIAnalyzer(
	model string,
) *OpenAIAnalyzer {
	return &OpenAIAnalyzer{
		client: openai.NewClient(),
		model:  model,
	}
}

func (analyzer *OpenAIAnalyzer) Analyze(
	ctx context.Context,
	incidentContext IncidentContext,
) (Analysis, error) {

	contextJSON, err :=
		json.MarshalIndent(
			incidentContext,
			"",
			"  ",
		)

	if err != nil {
		return Analysis{},
			fmt.Errorf(
				"marshal incident context: %w",
				err,
			)
	}

	prompt :=
		fmt.Sprintf(
			`You are an SRE incident assistant.

Analyze the following monitoring incident.

Important rules:
- Do not decide whether the service is DOWN or HEALTHY.
- Treat the runtime status and incident status as authoritative backend facts.
- Do not invent infrastructure details that are not present.
- Possible causes must be phrased as hypotheses, not certainties.
- Recommended actions should be concrete and operational.
- Keep the response concise.

Return ONLY valid JSON using exactly this structure:

{
  "summary": "short incident summary",
  "possibleCauses": [
    "possible cause 1"
  ],
  "recommendedActions": [
    "action 1"
  ]
}

Incident context:

%s`,
			string(contextJSON),
		)

	response, err :=
		analyzer.client.Responses.New(
			ctx,
			responses.ResponseNewParams{
				Model: analyzer.model,
				Input: responses.ResponseNewParamsInputUnion{
					OfString: openai.String(
						prompt,
					),
				},
			},
		)

	if err != nil {
		return Analysis{},
			fmt.Errorf(
				"request OpenAI incident analysis: %w",
				err,
			)
	}

	output :=
		response.OutputText()

	var analysis Analysis

	err =
		json.Unmarshal(
			[]byte(output),
			&analysis,
		)

	if err != nil {
		return Analysis{},
			fmt.Errorf(
				"parse OpenAI incident analysis: %w; output=%q",
				err,
				output,
			)
	}

	if analysis.Summary == "" {
		return Analysis{},
			fmt.Errorf(
				"OpenAI incident analysis returned empty summary",
			)
	}

	return analysis, nil
}
