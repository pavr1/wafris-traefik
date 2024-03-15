# Wafris traefik plugin 

## Usage

### Define the plugin in Static Configuration

Wafris plugin must be first defined in your traefik [static configuration][static]

[static]: https://doc.traefik.io/traefik/getting-started/configuration-overview/#the-static-configuration


There are three different, mutually exclusive (i.e. you can use only one at the same time), ways to define static configuration options in Traefik:
	1. In a configuration file
	2. In the command-line arguments
	3. As environment variables

### Static Config: YAML or TOML example

YAML Static configuration example:

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

`my-router` is the primary router defined by `loadbalancer.yml`.  It takes any request to http://demo.localhost/

`service-foo` is our name for the web app or website that sites behind traefik that you are routing visitors to.

`waf-plugin` is the arbitrary name of the middleware you are putting between inbound traffic and your web app.  We define `waf-plugin` as a wrapper around the official Wafris traefik plugin and all Wafris configuration is done here.


<img src='https://uptimer.expeditedsecurity.com/wafris-traefik' width='0' height='0'>
