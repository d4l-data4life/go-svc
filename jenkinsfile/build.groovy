#!groovy

@Library(value='pipeline-lib@v2.12.0', changelog=false) _

buildPipeline projectName: 'go-svc',
    dockerRegistryID: 'phdp',
    deployable: false
