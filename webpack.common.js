const path = require('path');

module.exports = {
    entry: {
        'menu':'./scripts/menu/index.ts',
        'room':'./scripts/room/index.ts',
        'chess':'./scripts/chess/index.ts',
    },
    module: {
        rules: [
            {
                test: /\.ts$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    resolve: {
        extensions: [".ts", ".js"],
    },
    output: {
        filename: '[name].bundle.js',
        path: path.resolve(__dirname, 'build/scripts'),
        clean: true,
    },
}