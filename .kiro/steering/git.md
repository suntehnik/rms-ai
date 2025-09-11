# General GIT guidelines

## You should use git flow for branch management

- WHEN you start a new task you SHOULD create a new reature branch from freshly pulled from remote main branch
- WHEN you finish the task you SHOULD run 'make test-fast'
- WHEN you see error messages during 'make test-fast' you SHOULD fix errors first
- WHEN you finished working on task you SHOULD commit your changes with comprehensive commit message explaining what has been done and push to remote
- WHEN you finished commit and push you SHOULD create a github pull request
- WHEN you created a github pull request you SHOULD watch all checks in this pull request
- YOU SHOULD MARK TASK COMPLETE IF AND ONLY IF ALL PR CHECKS ARE CLEAN
- WHEN you see error messages in gihub PR checks you SHOULD analyse the reason and fix