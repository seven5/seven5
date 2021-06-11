// note that this expects CWD to be the directory containing this file
const path = require('path');
const CopyPlugin = require('copy-webpack-plugin');
const ESLintPlugin = require('eslint-webpack-plugin');
var webpack = require('webpack');

module.exports = {
    mode: 'development',
    entry: {
        main: {
          import:['./app/demo.ts'],
          filename:'demo.js'
        },
        test:{
            import:['./test/test.ts'],
            filename: 'alltests.js',
        }
    },
    output: {
        path: path.resolve(__dirname, 'dist'),
    },
    devServer: {
        open: 'http://localhost:8888',
        port: 8888,
        contentBase: path.resolve(__dirname, 'dist'),
    },
    resolve: {
        extensions: ['.ts','.js'], //.js is required for libs in node_modules
    },
    module: {
        rules: [
            {
                test: /.ts$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            }
        ]
    },
    plugins: [
        new ESLintPlugin({
            extensions: ['ts'],
            files: ['src/', 'app/'],
            fix: true, // equivalent of --fix on command line
        }),
        new CopyPlugin({
            patterns: [
                { from: 'static' },
                {from:'app', to:'app'},
                {from:'app/static', to:'app'},
                {from:'test', to:'test'},
            ],
        }),
    ],
    //devtool: false,
    devtool: 'inline-source-map',
};
