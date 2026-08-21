'use strict';

function greeting(name) {
  return `Hello from ${name}!`;
}

if (require.main === module) {
  console.log(greeting('${{ values.name }}'));
}

module.exports = { greeting };
