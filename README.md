<p>
  <a href="https://fountains.tommasoamici.com.com/#gh-light-mode-only" target="_blank">
    <img src="./.github/banner-light.svg" alt="Fountain finder" width="350" height="70">
  </a>
  <a href="https://fountains.tommasoamici.com.com/#gh-dark-mode-only" target="_blank">
    <img src="./.github/banner-dark.svg" alt="Fountain finder" width="350" height="70">
  </a>
</p>

Are you ever out on a walk and suddenly you feel thirsty? With this small application
you can find water fountains near you.

Powered by OpenStreetMap.

## Contributing

The application requires Redis, Go, and Node.js to be built.

```sh
# vite server
yarn dev
# go server
make run
```

You can optionally install [air](https://github.com/cosmtrek/air) to enable live reload
of the Go server.

Redis is used to cache responses from the Overpass API, which can be slow sometimes.
