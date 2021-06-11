// we are doing an arbitrary decode in this file, so we have to turn this off
/* eslint-disable @typescript-eslint/no-explicit-any */

export interface decodeable {
  decode(): any;
  typename(): string;
}

function instanceofDecodeable(object: any): object is decodeable {
  if (typeof object !== 'object') {
    return false;
  }
  return 'decode' in object;
}

/// ////////////////
// Utility functions for decoding things, useful for debug printing
/// ///////////////

type empty = { [key: string]: unknown };

/**
 * Decode returns a version of the input object with every attribute found decoded
 * into its value.  Thus a structure {... {... {foo:someAttribute}...}...} will return
 * {... {... {foo:valueOfAttribute} ...} ...}  Everything else is returned as-is.
 * Strictly, it replaces 'decodeable' with their vaule.
 */
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
export const decodeValue = (a: any): any => {
  if (instanceofDecodeable(a)) {
    return a.decode();
  }
  if (
    typeof a === 'object' &&
    Object.prototype.toString.call(a) === '[object Array]'
  ) {
    // array case
    const result = [] as any[];
    const asArray = a as Array<any>;
    asArray.forEach((elem: any) => {
      result.push(decodeValue(elem));
    });
    return result;
  }
  if (typeof a === 'object') {
    // object (dict)
    const result: empty = {};
    const keys = Object.keys(a) as Array<string>;
    keys.forEach((k: string) => {
      result[k] = decodeValue(a[k]);
    });
    return result;
  }
  // it's a simple type
  return a;
};

/**
 * DecodeTypename returns a structure that is the same as the input structure with
 * every value replaced by it's typeof except attributes.  Attributes return
 * a value like SimpleAttribute<name>  where name is the underlying type.
 * This can be highly valuable for debugging.  Arrays are converted into
 * [ typename ] or [ empty ]  where typename is the type of the array elements.
 */
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
export const decodeTypename = (a: any): any => {
  if (instanceofDecodeable(a)) {
    return a.typename();
  }
  if (
    typeof a === 'object' &&
    Object.prototype.toString.call(a) === '[object Array]'
  ) {
    // array case
    const asArray = a as Array<any>;
    if (asArray.length === 0) {
      return '[ empty ]';
    }
    return `[ ${decodeTypename(asArray[0])} ]`;
  }
  if (typeof a === 'object') {
    // object (dict)
    let result = '{ ';
    const keys = Object.keys(a) as Array<string>;
    keys.forEach((k: string) => {
      result += `${k}:${decodeTypename(a[k])},`;
    });
    return `${result} }`;
  }
  // it's a simple type
  return typeof a;
};
