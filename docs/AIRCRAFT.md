# Custom Aircraft

SkyEye ships with an embedded encyclopedia of aircraft data, which the
controller uses for various GCI decisions.

You can extend this encyclopedia to add support for community mods, or new
modules which have not yet been added to SkyEye's embedded data.

## Enabling the Feature

Create a YAML file containing a list of one or more aircraft entries, then set the path to it
in the `aircraft-file` setting in SkyEye's configuration.

## Aircraft Properties

Each entry supports the following properties:

| Property | Required | Description |
| --- | --- | --- |
| `acmi_short_name` | Yes | The aircraft's `ShortName` in [Tacview/ACMI telemetry](https://raia-software-inc.gitbook.io/tacview/technical-documentation/acmi-telemetry-file-format). Must match exactly; case and punctuation matter. |
| `tags` | Yes | A list of tags describing the aircraft. See [Tags](#tags). |
| `nato_reporting_name` | No | The NATO reporting name, e.g. `Flanker`, `Fulcrum`. Not all aircraft have one. |
| `nickname` | No | A common nickname, e.g. `Warthog`, `Viper`, `Scooter`. Not all aircraft have one. |
| `official_name` | No | The official name, e.g. `Thunderbolt II`, `Fighting Falcon`, `Skyhawk`. Not all aircraft have one. |
| `platform_designation` | No | The platform designation, e.g. `A-10`, `F-16`, `A-4`. |
| `type_designation` | No | The specific type designation, e.g. `A-10C`, `F-16C`, `A-4E`. |
| `threat_radius_nm` | No | The threat radius in nautical miles. See [Threat radius](#threat-radius). |
| `fuel_provider` | No | The refueling method this aircraft provides as a tanker. See [Refueling](#refueling). |
| `fuel_receiver` | No | The refueling method this aircraft requires to take fuel. See [Refueling](#refueling). |

### Reporting Name

The name SkyEye uses to call out a contact on the radio is its **reporting name**. SkyEye derives it
from the properties above in this order of preference: `nato_reporting_name`, then `nickname`, then
`official_name`, then `platform_designation`. Every aircraft must set at least one of these four
properties so SkyEye has a name to call it.

### Tags

`tags` is a list of one or more of the following values:

- `fixed-wing` — a fixed-wing aircraft.
- `rotary-wing` — a helicopter.
- `fighter` — a fighter armed with air-to-air missiles.
- `attack` — an attack aircraft with self-defense air-to-air missiles.
- `unarmed` — an aircraft with no air-to-air missiles (transports, tankers, AWACS, etc.).

Every aircraft must have **exactly one** of `fixed-wing` or `rotary-wing`. The `fighter`, `attack`,
and `unarmed` tags affect how contacts are grouped and, when you do not set `threat_radius_nm`, the
default threat radius.

### Threat Radius

`threat_radius_nm` is the range, in nautical miles, at which SkyEye considers the aircraft a threat.
This property is optional. If you omit it, SkyEye picks a sensible default from the aircraft's tags.

If you want to set it explicitly, these values are good starting points:

- **15 NM** for aircraft armed with older semi-active radar missiles, or infrared missiles.
- **25 NM** for aircraft armed with newer semi-active missiles, or active radar missiles.
- **35 NM** for fast interceptors and advanced fighters.

### Refueling

`fuel_provider` and `fuel_receiver` describe aerial refueling and drive VECTOR TANKER matching. Each
accepts one of:

- `boom` — flying boom refueling.
- `probe-and-drogue` — probe-and-drogue (basket) refueling.

The two properties apply to different kinds of aircraft, so set only the one relevant to yours:

- **Tanker aircraft:** set `fuel_provider` to the method the tanker dispenses. Leave `fuel_receiver`
  unset. This is what marks the aircraft as a tanker for VECTOR TO TANKER commands.
- **Receiver aircraft:** set `fuel_receiver` to the method the aircraft takes fuel with, if any. Leave
  `fuel_provider` unset.

When a player asks for a vector to a tanker, SkyEye only sends them to a tanker whose `fuel_provider`
matches their own aircraft's `fuel_receiver`.

## Overriding Built-In Aircraft

If an entry's `acmi_short_name` matches an aircraft already in the built-in encyclopedia, your entry
replaces the built-in one. This lets you customize the data for an aircraft that already ships with SkyEye.

## Example

The example below adds the A-4 Skyhawk community mod. It is an attack aircraft that uses
probe-and-drogue refueling, and its ACMI short name is `A-4E-C`.


```yaml
- acmi_short_name: A-4E-C
  tags:
  - fixed-wing
  - attack
  platform_designation: A-4
  type_designation: A-4E
  official_name: Skyhawk
  nickname: Scooter
  threat_radius_nm: 15
  fuel_receiver: probe-and-drogue
```

With this file loaded, SkyEye recognizes the A-4, calls it a "Scooter" (the `nickname` is preferred
over the `official_name`), treats it as an attack aircraft with a 15 NM threat radius, and sends it
to probe-and-drogue tankers.
