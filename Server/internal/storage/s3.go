package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/marmyr/iagdbackup/internal/config"
)

func ConnectAws() *session.Session {
	Region := os.Getenv(config.Region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Region)},
	)

	if err != nil {
		panic(err)
	}

	return sess
}

// CharacterPrefix is where a user's character archives live. Everything under
// it belongs to that single user, so deleting the prefix is how we clean up after an account deletion.
func CharacterPrefix(userId config.UserId) string {
	return fmt.Sprintf("characters/%v/", userId)
}

// DeleteS3Prefix removes every object under prefix. Paginates and batches
// internally, so a user with many characters costs a handful of requests rather than one per object.
func DeleteS3Prefix(prefix string) error {
	bucket := os.Getenv(config.BucketName)
	if bucket == "" {
		return fmt.Errorf("%s is not set", config.BucketName)
	}

	svc := s3.New(ConnectAws())
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	return s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter)
}
