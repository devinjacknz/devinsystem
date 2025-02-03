const transformer = {
  process(sourceText) {
    return {
      code: sourceText.replace(/import\.meta\.env/g, '(typeof process !== "undefined" ? process.env : import.meta.env)')
    };
  },
};

module.exports = transformer;
