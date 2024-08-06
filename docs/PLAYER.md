# Player Guide

## Quickstart for Multiplayer

1. Set your in-game callsign to a simple callsign in the format `Callsign 1 yourname`.
1. Join a server that uses SRS and SkyEye.
1. Get airborne.
1. Tune to the server's SkyEye frequency in SRS.
1. Say "Anyface, Callsign 1, radio check" and see if the bot understands you.
1. Say "Anyface, Callsign 1, picture" to be told about the highest priority threats near you. 
1. Say "Anyface, Callsign 1, bogey" to get a bearing to the nearest threat.

## A Word of Warning

> _"Hello. DCS is full of bugs"_

For various reasons, aircraft in DCS are sometimes marked as dead when they are still very much alive and dangerous. These "zombie" aircraft do not appear in exported data or even on the in-game F10 map, but can still shoot you down. Sometimes this happens to _your_ aircraft, and you become invisible to many game systems and data exports!

GCI is only one source of data to help you build situational awareness. It must not be considered an all-seeing eye. It does not replace your onboard sensors and communication tools such as your eyeballs, radio and attack radar - it assists you.

> _"All software is garbage."_

This bot is not particularly intelligent. It is good at counting enemy aircraft, and at computing angles and distances. However, it is very dumb at understanding the tactical or strategic situation. A human is always going to be better than the bot. I hope the bot can help fill in when a human controller is not available, but temper your expectations.

> _"Keep your stick on the ice."_

This is a silly piece of software for a silly computer game. Don't take it too seriously. Remember to HAVE FUN!

## Choosing Your Callsign

You need a callsign to use SkyEye. SkyEye will make its best effort to figure out your callsign from your in-game name, but it works best if you have a name like `Mobius 1 | Reaper` (MOBIUS ONE) or `Hitman 11 | Monarch` (HITMAN ONE ONE).

That is:

1. A two or three syllable English word.
2. One to three digits.
3. A pipe character (`|`)
3. Your non-callsign username.

Your callsign should be unique within a server. If multiple players have the same callsign, SkyEye will respond but you may receive inconsistent information. Note that callsigns are normalized in capitalization and numbers - "WARDOG 14", "Wardog 14" and "Wardog 1 4" are all considered to be the same callsign. Numbers ar pronounced individually - "Spare 15" is prnounced "Spare One Five", not "Spare Fifteen".

Avoid:

* Names that contain brevity codewords, including "alpha", "radio", "bogey", "picture", "declare", "snaplock", "spiked", "bullseye".
* Names that are hard to distinguish, like "Spare"/"Spear", "Jester"/"Gesture", "Witch"/"Which". The bot will make a best effort, but may be less accurate.
* Names that aren't widely recognized words in common parlance, like "Razgriz" or "Beskar". The bot will make a best effort, but may be less accurate.
* Names in poor taste.

If your callsign doesn't follow this format, SkyEye makes a best effort to understand it while still applying its parser rules. A bare username like "Jeff" (with no numbers) may still work, but do not expect this to work reliably.

## Using Skyeye

### Using Your Voice

You can send a request to SkyEye by speaking on the SkyEye frequency in SRS. The format of the request is:

`GCI_CALLSIGN YOUR_CALLSIGN (...) REQUEST_TYPE (...) (REQUEST_ARGUMENTS...) (...)`

Where:

1. `GCI_CALLSIGN` is either the GCI's callsign or "Anyface" - either is fine.
2. `YOUR_CALLSIGN` is your chosen callsign, e.g. "Mobius One" or "Hitman One One"
3. `REQUEST_TYPE` is a keyword indicating what kind of request you're sending (discussed below)
4. `REQUEST_ARGUMENTS` are optional modifiers to the request (discussed below)

Example: "Anyface, Mobius One, spiked One Five Zero" is parsed as `GCI_CALLSIGN="Anyface", YOUR_CALLSIGN="mobius 1", REQUEST_TYPE=SPIKED, REQUEST_ARGUMENTS=[150]`

The `(...)` indicates that you can say extra words in those spots and SkyEye will do its best to figure out what you mean.

Example: "Anyface, Mobius One, Alpha Check" and "Anyface, Mobius One, good morning. Alpha Check bullseye" are both parsed to `GCI_CALLSIGN="Anyface", YOUR_CALLSIGN="mobius 1", REQUEST_TYPE=ALPHA_CHECK, REQUEST_ARGUMENTS=[]`.

Some types of requests require you to provide numeric arguments.

* Compass bearings must be given by speaking each digit individually. Say "Six Five" or "Zero Six Five", not "Sixty-Five."
* All other numbers should be given normally - "Seventeen", not "One Seven"
* Do not use ICAO pronunciation; pronounce numbers normally. Say "Three", "Five", "Nine", not "Tree", "Fife", "Niner".
* When providing bullseye coordinates, you may either say "bullseye" before the coordinates, or omit the word "bullseye". That is, both "Bullseye Zero Six Five, Ninety-Nine" and "Zero Six Five, Ninety-Nine" are acceptable.
* When providing bullseye coordinates, speak at a steady and measured pace with a slight p;ause between each number. Not too fast, not too slow. Don't mush your numbers together.

Tips:

* Think about what you want to say before you say it.
* Speak clearly at a measured pace, as if you were recording a vlog or talking to colleages in a meeting room. Speaking too quickly or excessively slowly can confuse the bot.
* If you misspeak, release your Push-to-Talk key and start over rather than trying to correct yourself.
* Avoid chatter on the SkyEye frequency. This may delay responses to actual requests.

### Using the Keyboard and Mouse

🚧 NOT YET IMPLEMENTED 🚧

## Available Requests

### RADIO CHECK

Keyword: `RADIO`

Function: The GCI will respond if they both see you on scope and heard you.

Use: Testing communication with the bot.

Examples:

```
MOBIUS 1: "Thunderhead Mobius One radio check"
THUNDERHEAD: "Mobius One, five by five."
```

```
HITMAN 11: "Galaxy Hitman One One how's my radio working?"
GALAXY: "Hitman One One, Lima Charlie" [LIMA CHARLIE meaning LOUD & CLEAR]
```

```
YELLOW 13: "Goliath Yellow One Three radio"
GOLIATH: "Yellow One Three, loud and clear"
```

### ALPHA CHECK

Keyword: `ALPHA`

Function: The GCI will check if they see you on scope and tell you your approximate current location in bullseye format.

Examples:

```
MOBIUS 1: "Thunderhead Mobius One alpha check"
THUNDERHEAD: "Mobius One, Thunderhead, contact, alpha check bullseye 010/122"
```

```
HITMAN 11: "Galaxy Hitman One One checking in as fragged, request alpha check bullseye"
GALAXY: "Hitman One One, Galaxy, contact, alpha check bullseye 144/28"
```

```
YELLOW 13: "Goliath Yellow One Three alpha"
GOLIATH: "Yellow One Three, Goliath, contact, alpha check bullseye 088/5"
```

Tips: You can use this to coarsely check your INS navigation system in an aircraft without GPS. It is accurate to within several miles (accounting for potential lag time between when the bot checks the scope and when the response is sent on the radio).

### BOGEY DOPE

Keyword: `BOGEY`

Function: The GCI will give you the Bearing, Range, Altitude and Aspect from your aircraft to the nearest air-to-air threat.

Arguments:

1. Filter (optional)

Examples:

```
MOBIUS 1: "Thunderhead Mobius One bogey dope"
THUNDERHEAD: "Mobius One, group threat BRAA 071/13, 17000, flank south, hostile, Flanker"
```

```
HITMAN 11: "Galaxy Hitman One One looking for a bogey - anything interesting?"
GALAXY: "Hitman One One, group threat BRAA 055/71, 22000, flank north, hostile, Tomcat"
```

```
YELLOW 13: "Goliath Yellow One Three bogey"
GOLIATH: "Yellow One Three, group threat BRAA 188/45, 8000, hot, hostile, Eagle"
```

### DECLARE

Keyword: `DECLARE`

Function: You provide the position of a radar contact on your scope. The GCI will look for contacts in that area and tell you if they are hostile, friendly, a furball (mixed) or clean (nothing on scope).

You can provide the position using either Bullseye or BRAA format .

Arguments:

1. Bullseye (bearing and distance) or BR (bearing and range) (required)
2. Altitude (optional)
3. Track direction (optional)

Providing the optional arguments can help the GCI distinguish between contacts. If there's a friendly at 5000 feet and a hostile at 25000 feet, you may get a FURBALL response if you only provide the bullseye, or a specific response if you also provide altitude.

Examples:

```
// TODO
```

### PICTURE

Keyword: `PICTURE`

Function: The GCI will look for threats near you and rank them by relative danger. It will tell you the total number of groups in your area, as well as detailed information on the three highest priority threats.

Arguments:

1. Filter (optional)

Examples:

```
MOBIUS 1: "Thunderhead Mobius One, picture"
THUNDERHEAD: "Thunderhead, 5 groups. Group bullseye 192/41, 21000, track south, bandit, Flanker. Group bullseye 178/32, 9000, track east, bandit, Frogfoot. Group bullseye 181/44, 20000, track northwest, bandit, Frogfoot."
```

```
HITMAN 11: "Galaxy Hitman One One how's the picture looking?"
GALAXY: "Hitman One One, 6 groups. Group bullseye 211/27, 18000, track northwest, bandit, Frogfoot. Group bullseye 226/12, 7000, track northwest, bandit, Fulcrum. Group bullseye 193/47, 36000, track northeast, bandit, Foxhound."
```

Tips:

* Repeat this call at regular intervals to maintain situational awareness.
* Air combat is highly complex and the threat ranking algorithm is imperfect. The GCI might omit a highly dangerous adversary from the response if it is slightly further away or at a lower altitude compared to other threats. Exercise caution!
* Be considerate of your allies on the channel. The response contains a great deal of useful information, but can occupy the channel for 20-30 seconds. 

### SNAPLOCK

Keyword: `SNAPLOCK`

Function: This is a faster form of DECLARE intended for use during a BVR timeline. You tell the GCI the BRA (bearing, range, altitude) of a threat on your radar scope. The GCI will look for a group in that area and response with information.

Arguments:

1. Bearing from you to the contact (required)
2. Range from you to the contact (required)
3. Altitude of the contact (required)

Examples:

```
MOBIUS 1: "Thunderhead Mobius One, snaplock one two five, ten, eight thousand"
THUNDERHEAD: "Mobius 1, threat group BRAA 125/10, 8000, hot, hostile, two contacts, Flanker."
```

### SPIKED

Keyword: `SPIKED`

Function: You tell the GCI the approximate bearing to an airborne threat on your Radar Warning Receiver (RWR). The GCI responds with information about the nearest potential source within a 30 degree cone in that direction.

Arguments:

1. Bearing to the airborne radar threat (required)

Examples:

```
MOBIUS 1: "Thunderhead Mobius One, spiked zero eight zero"
THUNDERHEAD: "Mobius One, spike range 35, 16000, flank northeast, hostile, single contact."
```

```
HITMAN 11: "Galaxy Hitman One One sees a spike at zero six zero"
GALAXY: "Hitman One One, spike range 45, 8000, hot, hostile, single contact."
```

```
YELLOW 13: "Goliath Yellow One Three spiked three six zero"
GOLIATH: "Yellow One Three, Goliath clean three six zero"
```

Tips:

* The accuracy of this call is imperfect. The information you receive is a best effort guess. The GCI may misidentify the actual source of the radar signal.

## Broadcast Calls

### SUNRISE

When the GCI controller comes online, it will announced that its services are available using the code word "SUNRISE".

If you hear this in the middle of a mission, it probably means the bot crashed and had to be restarted!

### PICTURE

🚧 NOT YET IMPLEMENTED 🚧

### THREAT

🚧 NOT YET IMPLEMENTED 🚧

### FADED
When the GCI controller sees a contact disappear from the radar scope for at least 30 seconds, it will announce the contact is FADED.

**This is not a confirmation that the contact has been destroyed!** In DCS, it is possible for aircraft to be marked dead while they are still alive and dangerous.

Example:

```
THUNDERHEAD: "Thunderhead, single contact faded Bullseye 146/123, track west, hostile, Flanker"
```
