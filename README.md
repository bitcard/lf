# LF: Fully Decentralized Fully Replicated Key/Value Store

*(c)2018-2019 [ZeroTier, Inc.](https://www.zerotier.com/)* 
*Licensed under the [GNU GPLv3](LICENSE.txt)*

**LF is currently beta software!**

## Introduction

LF (pronounced "aleph") is a fully decentralized fully replicated key/value store.

Fully decentralized means anyone can run a node without obtaining special permission and all nodes are effectively equal. Fully replicated means every node stores all data.

LF is built on a [directed acyclic graph (DAG)](https://en.wikipedia.org/wiki/Directed_acyclic_graph) data model that makes synchronization easy and allows many different security and conflict resolution strategies to be used. Proof of work is used to rate limit writes to the shared data store on public networks and as one thing that can be considered by conflict resolution and trust algorithms. See [DESIGN.md](doc/DESIGN.md) for details. (This document is under construction.)

The name LF comes from the short story [The Aleph](https://en.wikipedia.org/wiki/The_Aleph_%28short_story%29) by Jorge Luis Borges and the novel [Mona Lisa Overdrive](https://en.wikipedia.org/wiki/Mona_Lisa_Overdrive) by William Gibson. Borges' story involves a point in space that contains all other points, a fitting metaphor for a data store where every node stores everything. Gibson's novel features a sci-fi take on Borges' concept. At one point a character calls it the "LF" because "aleph" has been mis-heard as an acronym. We used LF because there's already an open source project called Aleph, it gives the command line client `lf` a short name, and because two levels of obscure literary reference recursion is cool.

### Why Does This Exist?

The purpose of LF is to provide for decentralized systems what the key/value store aspects of [etcd](https://github.com/etcd-io/etcd) and [consul](https://www.consul.io) provide in centrally managed environments, namely a fast reliable store for small but critical pieces of information.

Most decentralized systems rely on distributed hash tables (DHTs) like [Kademlia](https://en.wikipedia.org/wiki/Kademlia) to fill this role. This has been the standard approach since [magnet links](https://en.wikipedia.org/wiki/Magnet_URI_scheme) and similar DHT-based schemes were developed for open file sharing applications in the early 2000s.

For several years ZeroTier has researched ways to more completely decentralize our network. This drive is both economic and philosophical. We have multiple enterprise clients that want to minimize hard dependency on third party physical infrastructure and one that wants very reliable operation in very unreliable environments. We don't mind cutting our hosting costs either. Philosophically we began life as a "re-decentralize the Internet" open source effort and that's still very much in our DNA.

DHTs are simply not good enough for our needs. They're vulnerable to multiple types of network level denial of service and ["Sybil"](https://en.wikipedia.org/wiki/Sybil_attack) attacks, are slow, and require a very reliable network to maintain full access to the entire global data set. We built LF to provide an alternative that trades much higher local storage overhead (which is cheap) for continuous availability, high performance, and much stronger security guarantees.

### Features and Benefits

* Easy to use and easy to deploy.
* Fully decentralized network with no mandatory single points of control or failure.
* Versatile security model allowing user choice between different conflict resolution mechanisms that can be used alone or in combination with one another. These include local heuristics, proof of work "weight," elective trust of other nodes, and (eventually) certificates.
* Fast sub-second nearline queries against the entire global data set at all times.
* Flexible record lookup API allowing multiple nested keys and range queries against 64-bit ordinals associated with each key.
* Encrypted record keys and values for strong security and privacy despite full global data set replication. Order preserving encryption techinques are leveraged to allow range queries without revealing keys or exact ordinal values.
* Novel proof-of-knowledge mechanism makes it impossible to generate valid records identified by a key whose plain text is not known, increasing the difficulty of data set poisoning attacks by naive attackers.

### Limitations and Disadvantages

* LF is only useful for small bits of information that don't change very often like certificates, keys, IP addresses, names, etc.
* The [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem) trade-off is availability and partition-tolerance, meaning eventual consistency and no transactions.
* High storage and moderately high bandwidth requirements for full nodes make full node operation unsuitable for small resource constrained devices.
* Storage requirements grow over time in a manner not unlike a block chain. Fortunately [storage is getting cheaper over time too](https://www.backblaze.com/blog/hard-drive-cost-per-gigabyte/). LF is designed to allow partial nodes and discarding of some old data to migitate data set growth but these features are not implemented yet and likely won't be needed for years.

## Building and Running

LF works on Linux, Mac, and probably BSD. It won't work on Windows yet but porting shouldn't be too hard if anyone wants it. It's mostly written in Go (1.11+ required) with some C for performance critical bits. Building it is easy. The only non-Go dependency is a reasonably recent version of [SQLite](https://sqlite.org/) whose libraries and header files will need to be available on the system.

If you have recent Go, a C compiler, and SQLite just type `make` and it should build.

Once we get out of beta we'll probably provide some pre-built binaries as is the custom for Go projects.

## Getting Started

LF comes with a full node implementation and a command line client that can be used to create and query records and to some limited extent control nodes. All of these things exist in the same `lf` binary that gets built when you type `make`.

Once you build the binary just run it for help.

LF ships out of the box with its command line client configured to query `lf.zerotier.com`, a public network node operated by ZeroTier. That means you can try a simple query right away:

```text
$ ./lf get bad horse#
Bad Horse, Bad Horse                                      | bad#0 horse#0
Bad Horse, Bad Horse                                      | bad#0 horse#1
He rides across the nation, the thoroughbred of sin       | bad#0 horse#2
He got the application that you just sent in              | bad#0 horse#3
It needs evaluation, so let the games begin               | bad#0 horse#4
A heinous crime, a show of force                          | bad#0 horse#5
(a murder would be nice of course!)                       | bad#0 horse#6
Bad Horse, Bad Horse                                      | bad#0 horse#7
Bad Horse, He's Bad!                                      | bad#0 horse#8
The evil league of evil is watching so beware             | bad#0 horse#9
The grade that you receive'll be your last, we swear      | bad#0 horse#10
So make the bad horse gleeful, or he'll make you his mare | bad#0 horse#11
You're saddled up; there's no recourse                    | bad#0 horse#12
It's "hi-yo, silver!"                                     | bad#0 horse#13
Signed: Bad Horse.                                        | bad#0 horse#14
```

Don't forget the trailing hash sign on `horse#`. Drop the `./` if you put the binary somewhere in your path.

These are the lyrics to [Bad Horse](https://www.youtube.com/watch?v=VNhhz1yYk2U) from the musical [Dr. Horrible's Sing-Along Blog](https://drhorrible.com). Yes, we put that in the public data store. We hope it's fair use.

### Selectors, Ordinals, and Values

Record keys are cryptographic objects called *selectors* that are generated from plain text selector *names* and *ordinals*. The name of a selector is any text or binary string. The ordinal is an unsigned 64-bit integer. (If no ordinal is specified, it is zero.) Names can only be looked up directly but if you know a name you can ask for ranges of its ordinals.

Bad Horse is stored as a series of records with two selectors and with the ordinals in the second selector placing them in order. You can see them in the output above. Now try a few variations:

```text
$ ./lf get bad horse
Bad Horse, Bad Horse | bad#0 horse#0

$ ./lf get bad horse#2#10
He rides across the nation, the thoroughbred of sin  | bad#0 horse#2
He got the application that you just sent in         | bad#0 horse#3
It needs evaluation, so let the games begin          | bad#0 horse#4
A heinous crime, a show of force                     | bad#0 horse#5
(a murder would be nice of course!)                  | bad#0 horse#6
Bad Horse, Bad Horse                                 | bad#0 horse#7
Bad Horse, He's Bad!                                 | bad#0 horse#8
The evil league of evil is watching so beware        | bad#0 horse#9
The grade that you receive'll be your last, we swear | bad#0 horse#10
```

When you ask for `bad horse#` (in the very first example) the trailing hash is expanded to `#0#18446744073709551615`. That huge number is the maximum value of a 64-bit unsigned integer. Leaving off the trailing hash is equivalent to `#0` and gets only ordinal zero. Using `#2#10` asks for ordinals two through ten inclusive.

### First Come, First Serve!

If LF is open and decentralized, what happens if someone does this?

```text
$ ./lf set bad horse#0 'Good Horse, Good Horse'
```

A record will be created and published. Chances are nobody will see it.

The already existing Bad Horse records have three big things going for them:

1. They're older and therefore more work (in a proof of work sense) has been added to the DAG since their creation. Each record references older records and when it gets stitched into the DAG its work is transferred to its ancestors all the way back to the beginning of time, increasing a metric called their *weight*. A determined attacker willing to buy a lot of compute power could overcome this, but the more time passes the more costly this attack becomes.
2. They are going to have lower *local reputation* scores on current nodes, including the one you are likely querying. When a record arrives at a running node that collides with an existing fully synchronized and verified record, it gets flagged as suspect. Lower reptuation records also don't get picked to be links for new records, meaning they pick up weight more slowly and get replicated more slowly (if at all).
3. *Oracle nodes* will gossip about the impostor, but not the original. An oracle node is a node configured to generate *commentary*, special records that perform the dual function of adding work to the DAG and commenting on records these nodes find suspect. The commentary of all oracle nodes is available but users are free to decide whether or not to trust it.

But what does happen to lower weight or lower reputation records? Try this:

```text
$ ./lf -json get bad horse#0
```

The LF API returns results as an array of arrays. Each array in the first dimension represents a set of records that share the same set of selector keys. The array in the second dimension is these records sorted in descending order of trust according to whatever conflict resolution mechanism(s) are active. The default behavior of the command line client is to show only the "winner" for each set of selectors, but `-json` tells it to request and dump everything.

Since names are first come first serve, short names like `bad` aren't the sorts of names you'll want to use for the first selector for records in a production system. Good naming strategies include reverse-DNS-order names similar to Java class names (e.g. `com.zerotier...`), unique GUIDs, and random strings. The latter options are good for systems that don't want to advertise their keys globally and want to avoid making their records available to users not in-the-know. Just remember that [naming things is one of the two hard things in computing](https://www.martinfowler.com/bliki/TwoHardThings.html).

### Running a Full Node

Running a node on the public network is easy:

```text
$ ./lf node-start &
```

If you are a normal user the node will keep its files in `$HOME/.lf`. If you are root it will be `/var/lib/lf`. Obviously you'll want to set this up via *systemd* or some other process-supervising system to run it on a real server.

Nodes listen by default on P2P port 9908 and HTTP port 9980. These can be changed with command line options.

Here are some of the node-specific files you'll see in the node's home directory:

* `lf.pid`: PID of running node
* `node.log`: Node log output (`tail -f` this file to watch your node synchronize)
* `identity-secret.pem`: Secret ECC key for node's commentary and also for encryption of node-to-node P2P communications
* `authtoken.secret`: HTTP API bearer auth token.
* `peers.json`: Periodically updated to cache information about P2P peers of this node
* `genesis.lf`: Genesis records for the network this node participates in
* `node.db`: SQLite indexing and meta-data database
* `records.lf`: Flat data file containing all records in binary serialized format. This file will get quite large over time.
* `weights.b??`: Record proof of work weights with 96-bit weights striped across three memory mapped files
* `graph.bin`: Memory mapped record linkage graph data structure
* `wharrgarbl-table.bin`: Static table used by proof of work algorithm

The node will also modify `client.json` to add its own local (127.0.0.1) HTTP URL so that client queries will by default go to the local LF node.

Watch `node.log` after you start your server for the first time and you'll see it synchronizing with the network. This can take a while. Once the node is fully synchronized you should be able to make queries against any data.

A few caveats for running nodes:

* We recommend a 64-bit system with a bare minimum of 1gb RAM for full nodes. Full nodes usually use between 384mb and 1gb of RAM and may also use upwards of 1gb of virtual address space for memory mapped files. 32-bit systems may have issues with address space exhaustion.
* Don't locate the node's files on a network share (NFS, CIFS, VM-host mount, etc.) as LF makes heavy use of memory mapping and this does not always play well with network drives. It could be slow, unreliable, or might not work at all. Needless to say a cluster of LF nodes must all have their own storage. Pooling storage isn't possible and defeats the purpose anyway.
* Hard-stopping a node with `kill -9` or a hard system shutdown may corrupt the database. We plan to improve this in the future but right now the code is not very tolerant of this.

### LFFS

Full nodes have the capability to mount sections of the data store on the host as a filesystem using FUSE. This works on Linux and Macintosh systems as long as FUSE is installed (see [OSXFUSE](https://osxfuse.github.io) for Mac). This offers a very easy and fast way for other software to use LF as a shared state cache.

Right now LFFS is somewhat limited:

* This virtual filesystem is absolutely **not** suitable for things like databases or other things that do a lot of random writes and use advanced I/O features!
* Sizes are limited to 1mb, which is quite excessive for LF! Setting a lower limit in `mounts.json` is recommended.
* Hard links, special modes like setuid/setgid, named pipes, device nodes, and extended attributes are not supported.
* Filesystem permission and ownership changes are not supported and chmod/chown are silently ignored.
* Name length is limited to 511 bytes. This is for names within a directory. There is no limit to how many directories can be nested.
* Committing writes is slow if you must perform proof of work, and there's currently no way to cancel a commit in progress. Local writes will appear immediately but their propagation across the network might take a while especially if PoW is needed.
* Changes to `mounts.json` require a node restart (we'll add it to the API at some point).
* A single LF owner is used for all new files under a mount and there's no way to change this without remounting.
* Once a filename is claimed by an owner there is currently no way to transfer ownership.

Some of these limitations might get fixed in the future. Others are intrinsic to the system.

## Sol: The Public Network

LF ships with a default configuration for a public network we call *Sol* after our home solar system. This configuration contains a default node URL and default *genesis records* allowing nodes to automatically bootstrap themselves.

Sol permits record values up to 1024 bytes in size, though records that big take a lot of proof-of-work to publish. It also includes an *authorization certificate*. Normally records require proof of work to create. An authorization certificate allows this to be skipped if the record's owner is signed and this signature is included.

We intend to use LF (via Sol) to decentralize our root server infrastructure. Our roots service quite a few ZeroTier customers so that means they'll be forced to create quite a lot of (tiny) LF records. Doing proof of work for all those is possible but costly, so we stuffed a certificate into Sol's configuration to let us cheat and be special. Think of it as our "fee" for creating and maintaining LF and donating it as open source to the community.

LF is open source. It's possible to make your own LF networks and configure them however you like. You can even create private LF networks that *require* signatures for anyone to add records. These will be of interest to our enterprise customers with private air-gapped environments.

## Future Work

Our next order of business will be to implement LF as a data backend for ZeroTier roots. Once this happens there will be a release of ZeroTier's core product that will demote our root servers from being the exclusive top-level anchor points of the ZeroTier universe to co-equal root servers alongside any others that users happen to set up.

We'll also be exploring other uses of LF within our own product and service line and other things that can be done with it.

## Contributing

If you find a bug or want to contribute to LF, please feel free to open a pull request!
