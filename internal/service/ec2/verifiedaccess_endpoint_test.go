// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVerifiedAccessEndpoint_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_basic(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "application_domain"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "vpc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "example"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_domain_prefix", "example"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, "load-balancer"),
					resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
					resource.TestCheckResourceAttr(resourceName, "sse_specification.0.customer_managed_key_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancer_options.0.load_balancer_arn"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.0.protocol", "https"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_ids.0"),
					resource.TestCheckResourceAttrSet(resourceName, "verified_access_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_networkInterface(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_networkInterface(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "application_domain"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "vpc"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "example"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_domain_prefix", "example"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, "network-interface"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_options.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_options.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_options.0.protocol", "https"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_ids.0"),
					resource.TestCheckResourceAttrSet(resourceName, "verified_access_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_tags1(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate), acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_tags2(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate), acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_tags1(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate), acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_basic(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_policyDocument(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_policyBase(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_policyUpdate(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate), policyDoc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
				),
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_policyBase(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

// Verifies load balancer subnet ID's can be updated without a crash
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/39186
func testAccVerifiedAccessEndpoint_subnetIDs(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_subnetIDs(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.0.subnet_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_subnetIDsUpdate(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_options.0.subnet_ids.#", "2"),
				),
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_Cidr(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_Cidr(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_options.0.port_range.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_Cidr_Update(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cidr_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_options.0.port_range.#", "2"),
				),
			},
		},
	})
}

func testAccVerifiedAccessEndpoint_Rds(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessEndpoint
	resourceName := "aws_verifiedaccess_endpoint.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessEndpointConfig_Rds(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rds_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rds_options.0.port", "3306"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"endpoint_domain_prefix",
				},
			},
			{
				Config: testAccVerifiedAccessEndpointConfig_Rds_Update(rName, acctest.TLSPEMEscapeNewlines(key), acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rds_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rds_options.0.port", "6033"),
				),
			},
		},
	})
}

func testAccCheckVerifiedAccessEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_endpoint" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Endpoint %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckVerifiedAccessEndpointExists(ctx context.Context, n string, v *types.VerifiedAccessEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVerifiedAccessEndpointConfig_base(rName, key, certificate string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_acm_certificate" "test" {
  private_key      = "%[2]s"
  certificate_body = "%[3]s"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "TLS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_trust_provider" "test" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "oidc"

  oidc_options {
    authorization_endpoint = "https://example.com/authorization_endpoint"
    client_id              = "s6BhdRkqt3"
    client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
    issuer                 = "https://example.com"
    scope                  = "test"
    token_endpoint         = "https://example.com/token_endpoint"
    user_info_endpoint     = "https://example.com/user_info_endpoint"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance_trust_provider_attachment" "test" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.test.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}

resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_base_tcp(rName, key, certificate string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_acm_certificate" "test" {
  private_key      = "%[2]s"
  certificate_body = "%[3]s"

  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_trust_provider" "test" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "oidc"

  native_application_oidc_options {
    authorization_endpoint 		= "https://example.com/authorization_endpoint"
    client_id              		= "s6BhdRkqt3"
    client_secret          		= "7Fjfp0ZBr1KtDRbnfVdmIw"
    issuer                 		= "https://example.com"
    public_signing_key_endpoint = "https://example.com/signing_endpoint"
    scope                  		= "test"
    token_endpoint         		= "https://example.com/token_endpoint"
    user_info_endpoint     		= "https://example.com/user_info_endpoint"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_instance_trust_provider_attachment" "test" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.test.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}

resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance_trust_provider_attachment.test.verifiedaccess_instance_id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_basic(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "load-balancer"
  sse_specification {
    customer_managed_key_enabled = false
  }
  load_balancer_options {
    load_balancer_arn = aws_lb.test.arn
    port              = 443
    protocol          = "https"
    subnet_ids        = [for subnet in aws_subnet.test : subnet.id]
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_networkInterface(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.test.id
    port                 = 443
    protocol             = "https"
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}


`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_tags1(rName, key, certificate, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.test.id
    port                 = 443
    protocol             = "https"
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    %[4]q = %[5]q
  }
}


`, rName, key, certificate, tagKey1, tagValue1))
}

func testAccVerifiedAccessEndpointConfig_tags2(rName, key, certificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.test.id
    port                 = 443
    protocol             = "https"
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id
  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}

`, rName, key, certificate, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVerifiedAccessEndpointConfig_policyBase(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.test.id
    port                 = 443
    protocol             = "https"
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id
}
`)
}

func testAccVerifiedAccessEndpointConfig_policyUpdate(rName, key, certificate, policyDocument string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 1),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "network-interface"
  network_interface_options {
    network_interface_id = aws_network_interface.test.id
    port                 = 443
    protocol             = "https"
  }
  policy_document          = %[4]q
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id
}
`, rName, key, certificate, policyDocument))
}

func testAccVerifiedAccessEndpointConfig_subnetIDs(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 2),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "load-balancer"
  sse_specification {
    customer_managed_key_enabled = false
  }
  load_balancer_options {
    load_balancer_arn = aws_lb.test.arn
    port              = 443
    protocol          = "https"
    subnet_ids        = [for subnet in slice(aws_subnet.test, 0, 1) : subnet.id]
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_subnetIDsUpdate(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessEndpointConfig_base(rName, key, certificate, 2), fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" {
  application_domain     = "example.com"
  attachment_type        = "vpc"
  description            = "example"
  domain_certificate_arn = aws_acm_certificate.test.arn
  endpoint_domain_prefix = "example"
  endpoint_type          = "load-balancer"
  sse_specification {
    customer_managed_key_enabled = false
  }
  load_balancer_options {
    load_balancer_arn = aws_lb.test.arn
    port              = 443
    protocol          = "https"
    subnet_ids        = [for subnet in aws_subnet.test : subnet.id]
  }
  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_Cidr(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base_tcp(rName, key, certificate, 2),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" { 
  attachment_type        = "vpc"
  description            = "example"
  endpoint_type          = "cidr"
  sse_specification {
    customer_managed_key_enabled = false
  }
  cidr_options {
	cidr 			  = aws_subnet.test[0].cidr_block
	port_range {
	 from_port 		  = 443
	 to_port          = 443
    }
	protocol          = "tcp"
    subnet_ids        = [for subnet in aws_subnet.test : subnet.id]
  }

  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_Cidr_Update(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base_tcp(rName, key, certificate, 2),
		fmt.Sprintf(`
resource "aws_verifiedaccess_endpoint" "test" { 
  attachment_type        = "vpc"
  description            = "example"
  endpoint_type          = "cidr"
  sse_specification {
    customer_managed_key_enabled = false
  }
  cidr_options {
	cidr 			  = aws_subnet.test[0].cidr_block
	port_range {
	 from_port 		  = 443
	 to_port          = 443
    }
	port_range {
	 from_port 		  = 9443
	 to_port          = 9446
    }
	protocol          = "tcp"
    subnet_ids        = [for subnet in aws_subnet.test : subnet.id]
  }

  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_Rds(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base_tcp(rName, key, certificate, 2),
		fmt.Sprintf(`

# Security Group para permitir acceso a la BD solo desde la VPC
resource "aws_security_group" "testrds" {
  name        = "rds-security-group"
  description = "Grant rds access from VPC"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "rds-subnet-group"
  subnet_ids = [for subnet in aws_subnet.test : subnet.id]

  tags = {
    Name = "RDS Subnet Group"
  }
}

resource "aws_db_instance" "test" {
  allocated_storage      = 10                      
  engine                 = "mysql"                  
  engine_version         = "8.0"                    
  instance_class         = "db.t4g.micro"            
  identifier             = "basic-rds-instance"
  username               = "tfaccrds"               
  password               = "SuperSecure123!"         
  parameter_group_name   = "default.mysql8.0"
  publicly_accessible    = false                     
  skip_final_snapshot    = true                      
  storage_encrypted      = false                     
  multi_az               = false                     

  vpc_security_group_ids = [aws_security_group.testrds.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_verifiedaccess_endpoint" "test" {
  attachment_type        = "vpc"
  description            = "example"
  endpoint_type          = "rds"
  
  rds_options {
	port                 = aws_db_instance.test.port 
	instance_arn         = aws_db_instance.test.arn
    endpoint             = regex("^(.*):[0-9]+$", aws_db_instance.test.endpoint)[0]
    protocol          	 = "tcp"
    subnet_ids           = [for subnet in aws_subnet.test : subnet.id]
  }

  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}

func testAccVerifiedAccessEndpointConfig_Rds_Update(rName, key, certificate string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessEndpointConfig_base_tcp(rName, key, certificate, 2),
		fmt.Sprintf(`

# Security Group para permitir acceso a la BD solo desde la VPC
resource "aws_security_group" "testrds" {
  name        = "rds-security-group"
  description = "Grant rds access from VPC"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 6033
    to_port     = 6033
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "rds-subnet-group"
  subnet_ids = [for subnet in aws_subnet.test : subnet.id]

  tags = {
    Name = "RDS Subnet Group"
  }
}

resource "aws_db_instance" "test" {
  allocated_storage      = 10                      
  engine                 = "mysql"                  
  engine_version         = "8.0"                    
  instance_class         = "db.t4g.micro"            
  identifier             = "basic-rds-instance"
  username               = "tfaccrds"               
  password               = "SuperSecure123!"         
  parameter_group_name   = "default.mysql8.0"
  publicly_accessible    = false                     
  skip_final_snapshot    = true                      
  storage_encrypted      = false                     
  multi_az               = false

  vpc_security_group_ids = [aws_security_group.testrds.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
  port                   = 6033
}

resource "aws_verifiedaccess_endpoint" "test" {
  attachment_type        = "vpc"
  description            = "example"
  endpoint_type          = "rds"
  
  rds_options {
	port                 = aws_db_instance.test.port
	instance_arn         = aws_db_instance.test.arn
    endpoint             = regex("^(.*):[0-9]+$", aws_db_instance.test.endpoint)[0]
    protocol          	 = "tcp"
    subnet_ids           = [for subnet in aws_subnet.test : subnet.id]
  }

  security_group_ids       = [aws_security_group.test.id]
  verified_access_group_id = aws_verifiedaccess_group.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, key, certificate))
}
