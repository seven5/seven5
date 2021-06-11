module.exports = {
  extends: ['airbnb-typescript-prettier'],
  rules: {
    // we use namespaces to "wrap up and make available" our apis
    '@typescript-eslint/no-namespace': 0,
    // we use lower case names to indicate things that are intended to be used
    // by user level code
    'new-cap': 0,
    // we use _foo internally to mean that the property _foo is visible to user
    // code and likely has a get/set pair
    'no-underscore-dangle': ["error", {"allowAfterThis":true} ],
    '@typescript-eslint/no-unused-vars': ["error", { "varsIgnorePattern": "[iI]gnored" }],
    'no-param-reassign': ["error", { "props": false}],
    'import/no-extraneous-dependencies': ["error", {"devDependencies": ["**/jest/*.ts"]}]

  },
  // this horrible hack is because of an old (and still not fixed) bug in eslint's
  // react support.  if you don't do this, you'll get an annoying warning about react
  // versions, despite this not being a react project.
  settings:{
    react: {
      version: '999.999.999'
    }
  }
};
