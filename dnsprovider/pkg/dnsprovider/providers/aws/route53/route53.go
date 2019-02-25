/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// route53 is the implementation of pkg/dnsprovider interface for AWS Route53
package route53

import (
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

const (
	ProviderName = "aws-route53"
)

// MaxBatchSize is used to limit the max size of resource record changesets
var MaxBatchSize = 900

func init() {
	dnsprovider.RegisterDnsProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newRoute53(config)
	})
}

// route53HandlerLogger is a request handler for aws-sdk-go that logs route53 requests
func route53HandlerLogger(req *request.Request) {
	service := req.ClientInfo.ServiceName

	name := "?"
	if req.Operation != nil {
		name = req.Operation.Name
	}

	glog.V(4).Infof("AWS request: %s %s", service, name)
}

type route53Spec struct {
	AssumeRoleARN string `json:"assume-role-arn,omitempty"`
}

// newRoute53 creates a new instance of an AWS Route53 DNS Interface.
func newRoute53(config io.Reader) (*Interface, error) {
	var c route53Spec
	d := json.NewDecoder(config)
	if err := d.Decode(&c); err != nil {
		return nil, err
	}

	awsConfig := aws.NewConfig()
	// This avoids a confusing error message when we fail to get credentials
	// e.g. https://github.com/kubernetes/kops/issues/605
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true)

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	if c.AssumeRoleARN != "" {
		dnsCreds := stscreds.NewCredentials(sess, c.AssumeRoleARN, func(p *stscreds.AssumeRoleProvider) {
			p.RoleSessionName = "dnsprovider"
		})
		if _, err := dnsCreds.Get(); err != nil {
			return nil, err
		}
		dnsSess, err := session.NewSession(awsConfig.Copy(&aws.Config{Credentials: dnsCreds}))
		if err != nil {
			return nil, err
		}
		sess = dnsSess
	}

	svc := route53.New(sess)

	// Add our handler that will log requests
	svc.Handlers.Sign.PushFrontNamed(request.NamedHandler{
		Name: "k8s/logger",
		Fn:   route53HandlerLogger,
	})

	return New(svc), nil
}
