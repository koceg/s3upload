### s3upload `cli tool for aws s3 backup`

Goal of this tool is to provide a straight forward way of doing backups or saving arbitrary data in s3 buckets.

Current practice when you want to backup your database, the first thing you do is make a local copy of the db,table, etc that you are backing up and then use some tool that would save that data in the cloud.

Here I try to avoid that `extra step` and do a push/save directly to the s3 bucket.

### Requirements

```json
# custom policy
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:AbortMultipartUpload",
                "s3:ListMultipartUploadParts",
                "s3:ListBucketMultipartUploads"
            ],
            "Resource": "*"
        }
    ]
}
```

- go version go1.14.1 (should work with other versions as well that support `go mod`)
- AWS **access key ID / secret access key** required with the following **IAM** permissions:
  - AmazonS3 permissions
  	- use the custom policy above to create one and apply 
    - [Multipart upload API and permissions](https://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html) should provide additional information if errors arise

### Installation
```bash
> git clone https://github.com/koceg/s3upload.git
> cd s3upload
> go build
> sudo cp ./s3upload /usr/bin/s3upload # change permissions,ownership if neceserry
```
### Configuration
**~/.aws/config** file structure
```bash
[default]
region = <s3_bucket_region>
```

**~/.aws/credentials** file structure
```bash
[default]
aws_access_key_id = XXX
aws_secret_access_key = XXX
```
### Usage

If credentials are valid and `s3upload` is already in your working path all you have to do is **PIPE** the content to it and it will save it inside your s3 bucket 

```bash
s3upload -h # usage explanation

# simple test
> date -R | s3upload -u my_s3_bucket date
# would result in the following output
SUCCESS: https://s3.<bucket_region>.amazonaws.com/my_s3_bucket/YYYY/MM/DD/date_HH_MM_SS

# normal postgresql base backup of the server at mydbserver 
> pg_basebackup -h mydbserver -D /usr/local/pgsql/data

# would be rewritten as
> pg_basebackup -h mydbserver -D - | s3upload -u my_s3_bucket mydbserver

# OR backup and compress
> pg_dump -h localhost -U postgres -a -t table database | bzip2 | s3upload -u my_s3_bucket tabledata.bz2

# download object from s3 bucket
> s3upload -d <bucket> <object_key> <file_path> 
# Object Key would be everything after my_s3_bucket without forward slash see first example for reference

# dump object from s3 bucket to stdout and decompress
> s3upload -d <bucket> <object_key> | bzip2 -d
```
With that I think the use-case is clear and powerful.
