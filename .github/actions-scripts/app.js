import { Octokit } from "octokit";

const octokit = new Octokit({ 
  auth: process.env.TOKEN,
});

const [owner, repo] = process.env.GITHUB_REPOSITORY.split('/');
const pull_number = process.env.GITHUB_REF.split('/')[2];

const result = await octokit.rest.pulls.get({
    owner: owner,
    repo: repo,
    pull_number: pull_number,
    mediaType: {
      format: 'diff',
    },
  });
console.log(result.data)