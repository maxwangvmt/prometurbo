# turbo-goprobe-prometheus

This is a GO SDK probe that aims to discover applications and nodes from [Prometheus](https://prometheus.io/) for the
Turbonomic Operations Manager.  This probe is of **_prototype_** quality at the moment.

As of currently, this probe supports:
* Creating Application entities based on the Prometheus [istio](https://)
and the [redis](https://) exporters.  More will be gradually added in the future.
* Collecting app response time and transaction data.  More will be gradually added in the future.
* Stitching the discovered Application entities with their underlying Virtual Machine entities, provided that they are
discovered by the Turbo OpsMgr.

To try it out:

0. Prerequisites:
   * Install your Turbonomic OpsMgr.  The probe as of currently has been tested against version 6.2.
   * Install your Prometheus server and supported exporters (as listed above).
1. Configuration
   * Customize conf.json to point to your Turbo OpsMgr instance.
   * Customize conf.json to point to your Prometheus exporters.
2. Run `go install ./...` to build and install.
3. Start the probe: `./prometurbo`
4. Confirm in your OpsMgr that a target has been created as specified in `conf.json`.
5. Browse the Turbo UI for:
   * The discovered Application entities
   * The transaction metric graph of your app instances
   * The relationship between the Application and its underlying Virtual Machine - you may want to add a AWS target for
  example in your Turbo OpsMgr to discover the corresponding VM.
