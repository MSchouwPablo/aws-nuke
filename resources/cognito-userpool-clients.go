package resources

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CognitoUserPoolClientResource = "CognitoUserPoolClient"

func init() {
	registry.Register(&registry.Registration{
		Name:   CognitoUserPoolClientResource,
		Scope:  nuke.Account,
		Lister: &CognitoUserPoolClientLister{},
	})
}

type CognitoUserPoolClientLister struct{}

func (l *CognitoUserPoolClientLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cognitoidentityprovider.New(opts.Session)

	userPoolsLister := &CognitoUserPoolLister{}
	userPools, poolErr := userPoolsLister.List(ctx, o)
	if poolErr != nil {
		return nil, poolErr
	}

	resources := make([]resource.Resource, 0)

	for _, userPoolResource := range userPools {
		userPool, ok := userPoolResource.(*CognitoUserPool)
		if !ok {
			logrus.Errorf("Unable to case CognitoUserPool")
			continue
		}

		listParams := &cognitoidentityprovider.ListUserPoolClientsInput{
			UserPoolId: userPool.ID,
			MaxResults: aws.Int64(50),
		}

		for {
			output, err := svc.ListUserPoolClients(listParams)
			if err != nil {
				return nil, err
			}

			for _, client := range output.UserPoolClients {
				resources = append(resources, &CognitoUserPoolClient{
					svc:          svc,
					id:           client.ClientId,
					name:         client.ClientName,
					userPoolName: userPool.Name,
					userPoolID:   userPool.ID,
				})
			}

			if output.NextToken == nil {
				break
			}

			listParams.NextToken = output.NextToken
		}
	}

	return resources, nil
}

type CognitoUserPoolClient struct {
	svc          *cognitoidentityprovider.CognitoIdentityProvider
	name         *string
	id           *string
	userPoolName *string
	userPoolID   *string
}

func (p *CognitoUserPoolClient) Remove(_ context.Context) error {
	_, err := p.svc.DeleteUserPoolClient(&cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   p.id,
		UserPoolId: p.userPoolID,
	})

	return err
}

func (p *CognitoUserPoolClient) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", p.id)
	properties.Set("Name", p.name)
	properties.Set("UserPoolName", p.userPoolName)
	return properties
}

func (p *CognitoUserPoolClient) String() string {
	return fmt.Sprintf("%s -> %s", *p.userPoolName, *p.name)
}
