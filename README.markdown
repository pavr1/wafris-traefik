pvillalobos: please have in mind these are improvement areas and recommendations, it's up to the client to choose the ones that thinks are better for this tool.

# Wafris traefik plugin 
pvillalobos: a brief description of what this tool does can be put here.

## Prerequisits
pvillalobos: have a new part here to install any prerequisite 3rd party tool for developer purposes.
I think we will need to have 2 separate documents, one for internal developers (to give support to the plugin), and another one for clients (to follow up the steps to use the plugin)

## Usage

### Define the plugin in Static Configuration

Wafris plugin must be first defined in your traefik [static configuration][static]

[static]: https://doc.traefik.io/traefik/getting-started/configuration-overview/#the-static-configuration
There are three different, mutually exclusive (i.e. you can use only one at the same time), ways to define static configuration options in Traefik:
	1. In a configuration file
	2. In the command-line arguments
	3. As environment variables

pvillalobos: according to the documentation indeed there are 3 ways to set the plugin up. After certain tries to do it myself under traefik:v3.1 container, I got to the point where the plugin was successfully downloaded and installed locally but traefik was either crashing or not getting the plugin with no logs or limitted detailed ones. Not sure if clients have had problems installing this plugin like I did, in such case it would be good having some time to execute each installation and document every single step with good detail to make it easiear for clients.

pvillalobos: Another important thing I recommend is setting up a local environment with docker-composer to run this app locally. Any issue to support or new feature would be easier to handle and tested if a local environment is there with a simple command `docker-compose up`. So this
local environment would eventually have:
 - The plugin functionality
 - Redis image to connect against.
 - Traefik image with the 3 options to connect with the plugin.

pvillalobos: I see all files are stored in the same root directory, even though this is a small project we could expose the plugin itself and package the redis and proxy features, so they are not exposed.

### Static Config: YAML or TOML example

YAML Static configuration example:

pvillalobos: Maybe we can give more detail here, what's this configuration file's name (traefik.yaml), and where should the client put it.

```YAML
# Define the module name for the wafris plugin
# we use wafrisPlugin in this example, but any valid module name works
experimental:
  plugins:
    wafrisPlugin:
      moduleName: github.com/Wafris/wafris-traefik
      version: v0.0.1
```

TOML Static configuration example:

```TOML
# Define the module name for the wafris plugin
# we use wafrisPlugin in this example, but any valid module name works

experimental:
  plugins:
    wafrisPlugin:
      moduleName: github.com/Wafris/wafris-traefik
      version: v0.0.3
```

### Static Config: CLI example

In this example, we use the name wafrisPlugin.  Any valid module name should work.

```
--experimental.plugins.wafrisPlugin.modulename=github.com/Wafris/wafris-traefik --experimental.plugins.wafrisPlugin.version=v0.0.1
```

### Add the plugin to a provider or router

In your `traefik.yml` or equivalent file, you typically create a provider.  In this case we have an example provider defined by the `loadbalancer.yml` config file: 

```YAML
providers:
  # Enable the file provider to define routers / middlewares / services in file
  file:
    filename: loadbalancer.yml
```

The `loadbalancer.yml` config file can then be configured like so:

```YAML
http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - waf-plugin

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:2001
  
  middlewares:
    waf-plugin:
      plugin:
        wafrisPlugin:
          url: "redis://localhost:6379?protocol=3"
          wafris_timeout: 1.5
    
```

pvillalobos: we can give more details and examples here. The idea would be this document is the "go to" to install this plugin, clear and detailed.
`my-router` is the primary router defined by `loadbalancer.yml`.  It takes any request to http://demo.localhost/

`service-foo` is our name for the web app or website that sites behind traefik that you are routing visitors to.

`waf-plugin` is the arbitrary name of the middleware you are putting between inbound traffic and your web app.  We define `waf-plugin` as a wrapper around the official Wafris traefik plugin and all Wafris configuration is done here.


<img src='https://uptimer.expeditedsecurity.com/wafris-traefik' width='0' height='0'>
