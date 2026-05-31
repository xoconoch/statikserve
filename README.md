# Statikserve

Serve a static site publicly and upload new builds via a secret token.
Set the `AUTH_TOKEN` env variable in the compose file and push your static site as follows:

```bash
curl -X POST -H "Authorization: Bearer $AUTH_TOKEN" -F "file=@site.zip" http://localhost/_theres_no_way_you_have_this_in_your_static_site
```

The site.zip file shoudl contain the contents of the root of your site. For example, this one-liner works for quartz workflow:
```bash
rm -f site.zip && pnpm quartz build && (cd public && zip -r ../site.zip .) && curl -X POST "https://blog.cordovault.com/_theres_no_way_you_have_this_in_your_static_site" -H "Authorization: Bearer $STATIK" -F "file=@site.zip"
```
