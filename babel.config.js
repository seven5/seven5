// jest doesn't know about typescript other than to invoke babel to do
// transpilation.  so, we need this config file to tell babel what we want
// it to do.
module.exports = {
    presets: [
        ['@babel/preset-env', {targets: {node: 'current'}}],
        '@babel/preset-typescript',
    ],
};
