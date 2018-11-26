variable "name" {
  default = "file2s3hal"
}


resource "aws_s3_bucket" "cache" {
  bucket = "${var.name}-cache"
  acl = "private"
}

resource "aws_s3_bucket" "output" {
  bucket = "${var.name}"
  acl = "public-read"
}


resource "aws_s3_bucket" "build" {
  bucket = "${var.name}-build"
  acl = "private"
}

resource "aws_iam_role" "build-pipeline" {
  name = "build-${var.name}-pipeline"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role" "build-codebuild" {
  name = "build-${var.name}-app"
  assume_role_policy = <<EOF1
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF1
}

resource "aws_iam_role_policy" "build-policy" {
  name = "build-${var.name}"
  role = "${aws_iam_role.build-codebuild.name}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.cache.arn}",
        "${aws_s3_bucket.cache.arn}/*",
        "${aws_s3_bucket.build.arn}",
        "${aws_s3_bucket.build.arn}/*",
        "${aws_s3_bucket.output.arn}",
        "${aws_s3_bucket.output.arn}/*"


      ]
    },
    {
        "Effect": "Allow",
        "Action": [
           "codebuild:*"
        ],
        "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "build-pipeline" {
  name = "pipeline-${var.name}"
  role = "${aws_iam_role.build-pipeline.name}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.build.arn}",
        "${aws_s3_bucket.build.arn}/*"
      ]
    },
    {
        "Effect": "Allow",
        "Action": [
           "codebuild:*"
        ],
        "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_codebuild_project" "pipeline" {
  name = "${var.name}"
  service_role = "${aws_iam_role.build-codebuild.arn}"
  "artifacts" {
    type = "CODEPIPELINE"
  }
  "environment" {
    compute_type = "BUILD_GENERAL1_SMALL"
    image = "aws/codebuild/golang:1.10"
    type = "LINUX_CONTAINER"
  }
  "source" {
    type = "CODEPIPELINE"
  }
  cache {
    type = "S3"
    location = "${aws_s3_bucket.cache.bucket}"
  }
}

resource "aws_codepipeline" "web-app" {
  name = "${var.name}"
  "artifact_store" {
    location = "${aws_s3_bucket.build.bucket}"
    type = "S3"
  }
  role_arn = "${aws_iam_role.build-pipeline.arn}"
  "stage" {
    name = "source"
    "action" {
      category = "Source"
      name = "Source"
      owner = "ThirdParty"
      provider = "GitHub"
      version = "1"
      configuration {
        Owner = "weAutomateEverything"
        Repo = "file2s3hal"
        Branch = "master"
        OAuthToken = "${var.github_key}"
      }
      output_artifacts = [
        "source"]
    }
  },
  "stage" {
    name = "Build"
    "action" {
      input_artifacts = [
        "source"]
      output_artifacts = [
        "compiled"]
      category = "Build"
      name = "build"
      owner = "AWS"
      provider = "CodeBuild"
      version = "1"
      configuration {
        ProjectName = "${var.name}"
      }
    }
  }
}
