package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_code_repository", &resource.Sweeper{
		Name: "aws_sagemaker_code_repository",
		F:    testSweepSagemakerCodeRepositories,
	})
}

func testSweepSagemakerCodeRepositories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListCodeRepositoriesPages(&sagemaker.ListCodeRepositoriesInput{}, func(page *sagemaker.ListCodeRepositoriesOutput, lastPage bool) bool {
		for _, instance := range page.CodeRepositorySummaryList {
			name := aws.StringValue(instance.CodeRepositoryName)

			input := &sagemaker.DeleteCodeRepositoryInput{
				CodeRepositoryName: instance.CodeRepositoryName,
			}

			log.Printf("[INFO] Deleting SageMaker Code Repository: %s", name)
			if _, err := conn.DeleteCodeRepository(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Code Repository (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Code Repository sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Code Repositorys: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerCodeRepository_basic(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerCodeRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerCodeRepositoryExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/terraform-providers/terraform-provider-aws.git"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerCodeRepository_disappears(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerCodeRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerCodeRepositoryExists(resourceName, &notebook),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerCodeRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerCodeRepositoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_code_repository" {
			continue
		}

		describeNotebookInput := &sagemaker.DescribeCodeRepositoryInput{
			CodeRepositoryName: aws.String(rs.Primary.ID),
		}
		CodeRepository, err := conn.DescribeCodeRepository(describeNotebookInput)
		if err != nil {
			return nil
		}

		if aws.StringValue(CodeRepository.CodeRepositoryName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Code Repository %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerCodeRepositoryExists(n string, notebook *sagemaker.DescribeCodeRepositoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Code Repository ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		opts := &sagemaker.DescribeCodeRepositoryInput{
			CodeRepositoryName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeCodeRepository(opts)
		if err != nil {
			return err
		}

		*notebook = *resp

		return nil
	}
}

func testAccAWSSagemakerCodeRepositoryBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/terraform-providers/terraform-provider-aws.git"
  }
}
`, rName)
}
