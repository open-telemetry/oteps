# OpenTelemetry Sandbox

The OpenTelemetry Sandbox is a place under OpenTelemetry's governance where the community can collaborate and experiment on projects that still aren't mature enough to be accepted as part of the official OpenTelemetry project.

## Motivation

Over the history of OpenTelemetry, there have been situations where people came to our community proposing interesting ideas to be adopted. There have also been vendors offering code donations to the project, some of which are now mostly unmaintained.

As a possible solution to this, this OTEP proposes a new GitHub organization, [opentelemetry-sandbox](https://github.com/opentelemetry-sandbox). This organization will host projects until there's confidence that they have a healthy community behind them. They would also serve as a neutral place for the community to conduct experiments.

The advantage of a sandbox organization is that OpenTelemetry's governance can be used, including Governance Committee (GC) and Technical Committee (TC) intermediation of conflicts, making sure it’s an inclusive place for people to collaborate while keeping the reputation of the OpenTelemetry project as a whole untouched, given that it would be clear that OpenTelemetry doesn’t officially support projects within the sandbox.

There is a desire, but not an expectation, that projects will be moved from the sandbox as an official SIG or incorporated into an existing SIG. Realistically, we know that experiments might get dropped. There’s also no expectation that the OpenTelemetry project will provide resources to the sandbox project, like extra GitHub CI minutes or Zoom meeting rooms, although we might evaluate individual requests.

This OTEP is inspired by [CNCF’s sandbox projects](https://www.cncf.io/sandbox-projects/), but the process is significantly different.

## Internal details

### Acceptance criteria

A low barrier to entry would be desired for the sandbox. While the process can be refined based on our experience, the initial proposal for the process is the following:

1. Proposals should be written following the template below and have one TC and/or GC sponsor, who will regularly provide the TC and GC information about the state of the project.
2. Once a sponsor is found, the TC and GC will vote on accepting this new project on the Slack channel #opentelemetry-gc-tc.
    1. After one week, the voting closes automatically, with the proposal being accepted if it has received at least one :thumbs-up: (that of the sponsor, presumably).
    2. If at least one :thumbs-down: is given, or a TC/GC member has restrictions about the project but hasn’t given a :thumbs-down:, the voting continues until a majority is reached or the restrictions are cleared.
    3. The voting closes automatically once a simple majority of the TC/GC electorate has chosen one side.
3. Proponents should abide by OpenTelemetry’s Code of Conduct (currently the same as CNCF’s).
4. There’s no expectation that small sandbox projects will have regular calls, but there is an expectation that all decisions will be made in public and transparently.
5. Sandbox projects do NOT have the right to feature OpenTelemetry’s name on their websites.

Initially, there are three slots for sandbox projects available. The GC and TC might vote to increase this number based on the experience with the first projects.

#### Template

> Project name:
>
> Repository name:
>
> Problems the project will solve:
>
> Motivation for joining the sandbox:
>
> Zoom room requested?

##### Example

> Project name: OpenTelemetry Collector Community Distributions
>
> Repository name: opentelemetry-collector-distributions
>
> Problems the project will solve: The OpenTelemetry Collector Builder allows people to create their own distributions, and while the OpenTelemetry Collector project has no intentions (yet) on hosting other more specialized distributions, some community members are interested in providing those distributions, along with best practices on building and managing such distributions, especially around the CI/CD requirements.
>
> Motivation for joining the sandbox: I would love to have more community members to contribute with their own distributions. I would also appreciate broader help in keeping them in sync with the upstream Collector.
>
> Zoom room requested? No

### Periodic reports

On the second [GC/TC joint call](https://docs.google.com/document/d/1jylE5uZCKV9mrPw8Qrc5ExGyRVbBdqcbWPwni-hB5dE) of the calendar year (likely in February), the TC/GC sponsors for the sandbox projects MUST provide an update about their sponsored project. Should the sponsor not be able to join the call, a short written report MUST be provided as part of the meeting notes. The sponsor is not expected to write the report themselves, but rather, relay the report produced by the sandbox project to the group. The sponsor is encouraged to follow the sandbox project's developments and provide feedback, improving its chance of success.

### Exiting the sandbox

Projects in the sandbox are evaluated annually. The evaluation is done by the Governance and Technical Committees based on the report produced by the sandbox project and presented by the sponsor to the TC/GC, which will then decide on one of the following possible outcomes:

* continue as a sandbox project for another year
* incorporation as part of an existing SIG
* acceptance as a new SIG
* archival

### Further details

* A new GitHub user group will be created with the current members of the TC and GC as members. This group shall be the admin for all repositories in the organization.
* Project proponents are added as maintainers and encouraged to recruit other maintainers from the community.
* Code hosted under this organization is owned by the OpenTelemetry project and is under the governance of OTel’s Governance Committee.

## Trade-offs and mitigations

None.

## Prior art and alternatives

* One obvious alternative is to let users collaborate on their own accounts and organizations, as some are doing today. There are two problems with that: the first is that those projects usually lack clear neutral governance, so that external contributors aren't sure what's going to happen with their code contributions. The second problem is lack of visibility: most of the current experiments and initiatives aren't listed in the [OpenTelemetry's registry](https://opentelemetry.io/ecosystem/registry/), and being under an OpenTelemetry organization would make those projects more visible while still at early stages. There's also a certain feeling of legitimacy to the project when it's accepted as a sandbox project.
* Another alternative is to direct users to CNCF's sandbox, but at this moment, the queue of projects applying for that is huge and there's no expectations that a project applying today would be considered within the next few months. Besides, the CNCF Sandbox implies that projects are standalone projects, while OpenTelemetry Sandbox projects would often be incorporated as part of an existing SIG.

## Open questions

* None

## Future possibilities

N/A.
