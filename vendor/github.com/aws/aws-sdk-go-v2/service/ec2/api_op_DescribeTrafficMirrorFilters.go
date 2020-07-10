// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
)

type DescribeTrafficMirrorFiltersInput struct {
	_ struct{} `type:"structure"`

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	DryRun *bool `type:"boolean"`

	// One or more filters. The possible values are:
	//
	//    * description: The Traffic Mirror filter description.
	//
	//    * traffic-mirror-filter-id: The ID of the Traffic Mirror filter.
	Filters []Filter `locationName:"Filter" locationNameList:"Filter" type:"list"`

	// The maximum number of results to return with a single call. To retrieve the
	// remaining results, make another call with the returned nextToken value.
	MaxResults *int64 `min:"5" type:"integer"`

	// The token for the next page of results.
	NextToken *string `type:"string"`

	// The ID of the Traffic Mirror filter.
	TrafficMirrorFilterIds []string `locationName:"TrafficMirrorFilterId" locationNameList:"item" type:"list"`
}

// String returns the string representation
func (s DescribeTrafficMirrorFiltersInput) String() string {
	return awsutil.Prettify(s)
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DescribeTrafficMirrorFiltersInput) Validate() error {
	invalidParams := aws.ErrInvalidParams{Context: "DescribeTrafficMirrorFiltersInput"}
	if s.MaxResults != nil && *s.MaxResults < 5 {
		invalidParams.Add(aws.NewErrParamMinValue("MaxResults", 5))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

type DescribeTrafficMirrorFiltersOutput struct {
	_ struct{} `type:"structure"`

	// The token to use to retrieve the next page of results. The value is null
	// when there are no more results to return.
	NextToken *string `locationName:"nextToken" type:"string"`

	// Information about one or more Traffic Mirror filters.
	TrafficMirrorFilters []TrafficMirrorFilter `locationName:"trafficMirrorFilterSet" locationNameList:"item" type:"list"`
}

// String returns the string representation
func (s DescribeTrafficMirrorFiltersOutput) String() string {
	return awsutil.Prettify(s)
}

const opDescribeTrafficMirrorFilters = "DescribeTrafficMirrorFilters"

// DescribeTrafficMirrorFiltersRequest returns a request value for making API operation for
// Amazon Elastic Compute Cloud.
//
// Describes one or more Traffic Mirror filters.
//
//    // Example sending a request using DescribeTrafficMirrorFiltersRequest.
//    req := client.DescribeTrafficMirrorFiltersRequest(params)
//    resp, err := req.Send(context.TODO())
//    if err == nil {
//        fmt.Println(resp)
//    }
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/DescribeTrafficMirrorFilters
func (c *Client) DescribeTrafficMirrorFiltersRequest(input *DescribeTrafficMirrorFiltersInput) DescribeTrafficMirrorFiltersRequest {
	op := &aws.Operation{
		Name:       opDescribeTrafficMirrorFilters,
		HTTPMethod: "POST",
		HTTPPath:   "/",
		Paginator: &aws.Paginator{
			InputTokens:     []string{"NextToken"},
			OutputTokens:    []string{"NextToken"},
			LimitToken:      "MaxResults",
			TruncationToken: "",
		},
	}

	if input == nil {
		input = &DescribeTrafficMirrorFiltersInput{}
	}

	req := c.newRequest(op, input, &DescribeTrafficMirrorFiltersOutput{})

	return DescribeTrafficMirrorFiltersRequest{Request: req, Input: input, Copy: c.DescribeTrafficMirrorFiltersRequest}
}

// DescribeTrafficMirrorFiltersRequest is the request type for the
// DescribeTrafficMirrorFilters API operation.
type DescribeTrafficMirrorFiltersRequest struct {
	*aws.Request
	Input *DescribeTrafficMirrorFiltersInput
	Copy  func(*DescribeTrafficMirrorFiltersInput) DescribeTrafficMirrorFiltersRequest
}

// Send marshals and sends the DescribeTrafficMirrorFilters API request.
func (r DescribeTrafficMirrorFiltersRequest) Send(ctx context.Context) (*DescribeTrafficMirrorFiltersResponse, error) {
	r.Request.SetContext(ctx)
	err := r.Request.Send()
	if err != nil {
		return nil, err
	}

	resp := &DescribeTrafficMirrorFiltersResponse{
		DescribeTrafficMirrorFiltersOutput: r.Request.Data.(*DescribeTrafficMirrorFiltersOutput),
		response:                           &aws.Response{Request: r.Request},
	}

	return resp, nil
}

// NewDescribeTrafficMirrorFiltersRequestPaginator returns a paginator for DescribeTrafficMirrorFilters.
// Use Next method to get the next page, and CurrentPage to get the current
// response page from the paginator. Next will return false, if there are
// no more pages, or an error was encountered.
//
// Note: This operation can generate multiple requests to a service.
//
//   // Example iterating over pages.
//   req := client.DescribeTrafficMirrorFiltersRequest(input)
//   p := ec2.NewDescribeTrafficMirrorFiltersRequestPaginator(req)
//
//   for p.Next(context.TODO()) {
//       page := p.CurrentPage()
//   }
//
//   if err := p.Err(); err != nil {
//       return err
//   }
//
func NewDescribeTrafficMirrorFiltersPaginator(req DescribeTrafficMirrorFiltersRequest) DescribeTrafficMirrorFiltersPaginator {
	return DescribeTrafficMirrorFiltersPaginator{
		Pager: aws.Pager{
			NewRequest: func(ctx context.Context) (*aws.Request, error) {
				var inCpy *DescribeTrafficMirrorFiltersInput
				if req.Input != nil {
					tmp := *req.Input
					inCpy = &tmp
				}

				newReq := req.Copy(inCpy)
				newReq.SetContext(ctx)
				return newReq.Request, nil
			},
		},
	}
}

// DescribeTrafficMirrorFiltersPaginator is used to paginate the request. This can be done by
// calling Next and CurrentPage.
type DescribeTrafficMirrorFiltersPaginator struct {
	aws.Pager
}

func (p *DescribeTrafficMirrorFiltersPaginator) CurrentPage() *DescribeTrafficMirrorFiltersOutput {
	return p.Pager.CurrentPage().(*DescribeTrafficMirrorFiltersOutput)
}

// DescribeTrafficMirrorFiltersResponse is the response type for the
// DescribeTrafficMirrorFilters API operation.
type DescribeTrafficMirrorFiltersResponse struct {
	*DescribeTrafficMirrorFiltersOutput

	response *aws.Response
}

// SDKResponseMetdata returns the response metadata for the
// DescribeTrafficMirrorFilters request.
func (r *DescribeTrafficMirrorFiltersResponse) SDKResponseMetdata() *aws.Response {
	return r.response
}