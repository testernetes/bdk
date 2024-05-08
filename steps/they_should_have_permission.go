package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var GroupHavePermission = stepdef.StepDefinition{
	Name: "group-has-permission",
	Text: "^group {text} should be able to {verb} {resource} in {namespace}$",
	Help: `Checks that a group has permissions`,
	Examples: `
	Then group atlas-admins should be able to create pods in default`,
	StepArg:  stepdef.CreateOptions,
	Function: GroupHavePermissionFunc,
}

var GroupHavePermissionFunc = func(ctx context.Context, t *stepdef.T, group, namespace string) error {
	t.Clientset.AuthorizationV1().
		LocalSubjectAccessReviews(namespace).
		Create(ctx, &authzv1.LocalSubjectAccessReview{
			Spec: authzv1.SubjectAccessReviewSpec{
				Groups: []string{group},
				ResourceAttributes: &authzv1.ResourceAttributes{
					Verb:     "get",
					Group:    "",
					Version:  "*",
					Resource: "pods",
				},
			},
		}, metav1.CreateOptions{})

	return nil
}
