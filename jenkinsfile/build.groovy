#!groovy

@Library(value='pipeline-lib@v2', changelog=false) _

buildPipeline projectName: 'go-svc',
    dockerRegistryID: 'phdp',
    deployable: false
