# jsonnet --yaml-stream .drone.jsonnet > .drone.yml

local test_in_go(go_ver) = {
  kind: 'pipeline',
  name: 'golang%s' % go_ver,

  steps: [
    {
      image: 'golangci/golangci-lint:latest',
      name: 'lint',
      commands: [
        'golangci-lint run',
      ],
    },
    {
      image: 'golang:%s' % go_ver,
      name: 'test',
      commands: [
        'go test -coverprofile=coverage.txt -covermode=atomic ./...',
      ],
    },
    {
      image: 'golang:%s' % go_ver,
      name: 'coverage',
      failure: 'ignore',
      commands: [
        # https://docs.codecov.io/docs/about-the-codecov-bash-uploader
        # https://docs.codecov.io/docs/supported-ci-providers
        'curl -s https://codecov.io/bash | bash',
      ],
      environment: {
        CODECOV_TOKEN: {
          from_secret: 'codecov-upload-token',
        },
      },
    },
  ],
};

std.map(test_in_go, ['1.13']) + [
  {
    kind: 'secret',
    name: 'codecov-upload-token',
    data: 'c4Ktn+MGi5VipkWZ0YKTKqKDuPDs3zGAR3TygYqfKlhOMBSpxt13Sx32El0z+BvItFQufwD7gfEQltYR+YUDCw==',
  },
]
