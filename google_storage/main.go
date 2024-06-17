package google_storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	gs "cloud.google.com/go/storage"
	storage "github.com/tomekwlod/go-storager"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	once          sync.Once
	googleStorage *GoogleStorage
)

type GoogleStorage struct {
	ctx        context.Context
	client     *gs.Client
	bucket     *gs.BucketHandle
	bucketname string
}

func new(ctx context.Context, b64JsonKey, bucketname string) error {
	// {
	// // if for some reason you have to use JSON but base64 encoded you can use this solution:
	// b, err := base64.StdEncoding.DecodeString(b64JsonKey)
	// if err != nil {
	// 	return err
	// }
	// client, err := gs.NewClient(ctx, option.WithCredentialsJSON(b))
	// }
	client, err := gs.NewClient(ctx, option.WithCredentialsJSON([]byte(b64JsonKey)))

	if err != nil {
		return err
	}

	// Setup client bucket to work from
	bucket := client.Bucket(bucketname)

	gs := &GoogleStorage{
		ctx:        ctx,
		client:     client,
		bucket:     bucket,
		bucketname: bucketname,
	}

	googleStorage = gs

	return nil
}

func Get() (*GoogleStorage, error) {
	if googleStorage == nil {
		return nil, errors.New("cannot get GoogleStorage instance as it hasn't been initialized yet; use Setup() first")
	}

	return googleStorage, nil
}

func Setup(ctx context.Context, b64JsonKey, bucketname string) *GoogleStorage {
	if b64JsonKey == "" || bucketname == "" {
		panic("b64JsonKey and bucketname are required")
	}

	once.Do(func() {
		err := new(ctx, b64JsonKey, bucketname)
		if err != nil {
			// TODO: not sure if panic is a good one here, but this runs on server init only so it's not that horrible... let's find alternatives though
			panic(err)
		}
	})
	return googleStorage
}

func (g *GoogleStorage) Close() error {
	return g.client.Close()
}

func (g GoogleStorage) Upload(file io.Reader, filename, contentType string) (*storage.File, error) {
	// TODO: streaming
	// https://github.com/GoogleCloudPlatform/golang-samples/blob/main/storage/objects/stream_file_upload.go

	filename = strings.TrimLeft(path.Clean(filename), "/")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	w := g.client.Bucket(g.bucketname).Object(filename).NewWriter(ctx)
	w.ChunkSize = 0
	w.ContentType = contentType

	if _, err := io.Copy(w, file); err != nil {
		return nil, fmt.Errorf("failed to copy to bucket: %v", err)
	}
	// Data can continue to be added to the file until the writer is closed.
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close: %v", err)
	}

	return &storage.File{
		PublicURL:  fmt.Sprintf("https://gs.googleapis.com/%s/%s", g.bucketname, filename), // there should be no public url for the restricted buckets
		StorageURL: fmt.Sprintf("gs://%s/%s", g.bucketname, filename),
		// TODO: Created
		// TODO: Size
	}, nil
}

// https://github.com/GoogleCloudPlatform/golang-samples/blob/main/storage/objects/get_metadata.go#L28
func (g GoogleStorage) getMetadata(path string) (*gs.ObjectAttrs, error) {
	o := g.client.Bucket(g.bucketname).Object(path)
	attrs, err := o.Attrs(g.ctx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Bucket: %v\n", attrs.Bucket)
	fmt.Printf("CacheControl: %v\n", attrs.CacheControl)
	fmt.Printf("ContentDisposition: %v\n", attrs.ContentDisposition)
	fmt.Printf("ContentEncoding: %v\n", attrs.ContentEncoding)
	fmt.Printf("ContentLanguage: %v\n", attrs.ContentLanguage)
	fmt.Printf("ContentType: %v\n", attrs.ContentType)
	fmt.Printf("Crc32c: %v\n", attrs.CRC32C)
	fmt.Printf("Generation: %v\n", attrs.Generation)
	fmt.Printf("KmsKeyName: %v\n", attrs.KMSKeyName)
	fmt.Printf("Md5Hash: %v\n", attrs.MD5)
	fmt.Printf("MediaLink: %v\n", attrs.MediaLink)
	fmt.Printf("Metageneration: %v\n", attrs.Metageneration)
	fmt.Printf("Name: %v\n", attrs.Name)
	fmt.Printf("Size: %v\n", attrs.Size)
	fmt.Printf("StorageClass: %v\n", attrs.StorageClass)
	fmt.Printf("TimeCreated: %v\n", attrs.Created)
	fmt.Printf("Updated: %v\n", attrs.Updated)
	fmt.Printf("Event-based hold enabled? %t\n", attrs.EventBasedHold)
	fmt.Printf("Temporary hold enabled? %t\n", attrs.TemporaryHold)
	fmt.Printf("Retention expiration time %v\n", attrs.RetentionExpirationTime)
	fmt.Printf("Custom time %v\n", attrs.CustomTime)
	fmt.Printf("\n\nMetadata\n")
	for key, value := range attrs.Metadata {
		fmt.Printf("\t%v = %v\n", key, value)
	}

	return attrs, nil
}

// https://github.com/GoogleCloudPlatform/golang-samples/blob/main/storage/objects/list_files.go
func (g GoogleStorage) List(path string) ([]*storage.File, error) {
	// TODO: ctx: https://github.com/GoogleCloudPlatform/golang-samples/blob/main/storage/objects/delete_file.go#L38C2-L38C5
	ctx := context.Background()

	it := g.client.Bucket(g.bucketname).Objects(ctx, &gs.Query{
		Prefix: path,
	})

	files := []*storage.File{}

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		files = append(files, &storage.File{
			Path:       attrs.Name,
			PublicURL:  fmt.Sprintf("https://gs.googleapis.com/%s/%s", g.bucketname, attrs.Name), // there should be no public url for the restricted buckets
			StorageURL: fmt.Sprintf("gs://%s/%s", g.bucketname, attrs.Name),
		})
	}
	return files, nil
}

// https://github.com/GoogleCloudPlatform/golang-samples/blob/main/storage/objects/delete_file.go
func (g GoogleStorage) Delete(object string) error {
	ctx := context.Background()

	obj := g.client.Bucket(g.bucketname).Object(object)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to delete the file is aborted
	// if the object's generation number does not match your precondition.
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("object.Attrs: %w", err)
	}
	obj = obj.If(gs.Conditions{GenerationMatch: attrs.Generation})

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("deleting google storage object `%q` couldn't be done: %w", object, err)
	}

	return nil
}
