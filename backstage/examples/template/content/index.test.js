'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const { greeting } = require('./index.js');

test('greeting includes the component name', () => {
  assert.equal(greeting('demo'), 'Hello from demo!');
});

test('greeting never returns an empty prefix', () => {
  assert.ok(greeting('x').startsWith('Hello from '));
});
