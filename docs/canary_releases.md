# Canary releases

There is a need to use canary releases in order to make sure newer versions
of the gathering conditions don't break the Insights Operator gatherer.

This is how a canary release would look like:

```mermaid
sequenceDiagram
    actor RT as Rules Team
    actor PT as Processing Team
    participant GC as Gathering Conditions
    participant AI as App Interface
    participant GCS as Gathering Conditions Service
    participant Unleash
    participant Monitoring
    actor C as Cluster
    actor CC as Canary Cluster

    Note left of GCS: Stable version is 1.0.1<br/>CANARY_VERSION = none<br/>STABLE_VERSION = 1.0.1
    Note left of Unleash: Canary population is 0%
    GCS -->> C: Get version 1.0.1
    GCS -->> CC: Get version 1.0.1

    RT ->> GC: Push a new tag 1.0.2
    RT ->> AI: Set the canary and stable versions<br/>CANARY_VERSION = 1.0.2<br/>STABLE_VERSION = 1.0.1
    PT ->> AI: Review and merge the PR
    AI ->> GCS: Update prod deployment
    
    rect rgb(250,128,114)
        alt canary or stable version not found
            GCS ->> GCS: CrashLoopBackOff
            PT ->> AI: Revert the PR
        end
    end

    GCS -->> C: Get version 1.0.1
    GCS -->> CC: Still get version 1.0.1

    RT ->> Unleash: Increase the canary population to X%
    GCS -->> C: Get version 1.0.1
    rect rgb(143,188,143)
        GCS -->> CC: Get version 1.0.2
    end
    
    RT ->> Monitoring: Analyze the canary archives and check they are generated well

    alt canary archives are wrong
        RT ->> AI: Revert the PR
        PT ->> AI: Review and merge the revert
        AI ->> GCS: Update prod deployment
        GCS -->> C: Get version 1.0.1
        GCS -->> CC: Get version 1.0.1
    else canary archives look fine
        RT ->> AI: Remove the canary version and set stable <br/>CANARY_VERSION = none<br/>STABLE_VERSION = 1.0.2
        PT ->> AI: Review and merge the PR
        AI ->> GCS: Update prod deployment
        GCS -->> C: Get version 1.0.2
        GCS -->> CC: Get version 1.0.2
        RT ->> Unleash: Set the canary population to 0%
    end
```

In case of an emergency release, we would just need to create the app-interface
Pull Request and ignore the Unleash settings.

In case there are new conditions published during a canary release, you can
update the `CANARY_VERSION` and start the sequence from the beginning.

Note that the stage environment is not mentioned in the whole diagram. The
idea is that the service logic is tested both in stage and production, but the
new conditions are shipped to production directly via environment variables. The
service has to download the conditions on startup and fail if there is any issue,
so that we can see the Crashloopback and revert the PR. As Kubernetes waits for
the new pods to be healthy before deleting the old ones, a corrupted conditions
file shouldn't affect any client.
