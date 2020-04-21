package amazon

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/docker/ecs-plugin/pkg/compose"
)

func (c *client) ComposeUp(project *compose.Project, loadBalancerArn *string) error {
	template, err := c.Convert(project, loadBalancerArn)
	if err != nil {
		return err
	}

	json, err := template.JSON()
	if err != nil {
		return err
	}

	_, err = c.CF.ValidateTemplate(&cloudformation.ValidateTemplateInput{
		TemplateBody: aws.String(string(json)),
	})
	if err != nil {
		return err
	}

	_, err = c.CF.CreateStack(&cloudformation.CreateStackInput{
		OnFailure:        aws.String("DELETE"),
		StackName:        aws.String(project.Name),
		TemplateBody:     aws.String(string(json)),
		TimeoutInMinutes: aws.Int64(10),
	})
	if err != nil {
		return err
	}

	events, err := c.CF.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(project.Name),
	})
	if err != nil {
		return err
	}
	for _, event := range events.StackEvents {
		fmt.Printf("%s %s\n", *event.LogicalResourceId, *event.ResourceStatus)
		if *event.ResourceStatus == "CREATE_FAILED" {
			fmt.Fprintln(os.Stderr, event.ResourceStatusReason)
		}
	}

	// TODO monitor progress
	return nil
}