# Statikserve

Serve a static site publicly and upload new builds via a secret token.
Set the `AUTH_TOKEN` env variable in the compose file and push your static site as follows:

```bash
curl -X POST -H "Authorization: Bearer $AUTH_TOKEN" -F "file=@site.zip" http://localhost/_theres_no_way_you_have_this_in_your_static_site
```

The site.zip file should contain the dist/ dir. So generating it should look
something like `zip -r site.zip dist/`
