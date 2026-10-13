# Exact-channel rebalance screenshot fixtures

These images render the real `RebalanceChannelDialogContent` component against
a local, non-paying fixture. They contain synthetic identifiers only:

- peer keys use repeated `a`, `b`, `c`, or `d` characters;
- channel IDs use repeated `1` or `2` characters;
- balances, quote IDs, fingerprints, SCIDs, and timestamps are fixed fixtures;
- no Hub wallet, database, API credential, invoice, route, or payment record was
  read to create them.

The 320, 375, 768, and 1280 files cover the pre-quote and quote-review states.
The confirmation files show the exact typed-confirmation gate without invoking
the execution action. The executing, succeeded, failed, and recovery files show
synthetic saved-operation records; no execution request was made.

## SHA-256

```text
8ec6860a30eb2782dc4ac272f883289c1aad02105e6e82a96d787f6bec3a7f90  confirmation-1280.jpg
1699608872426da2474a4aa55936f2540e16dd7911f914f1bde214f08d702e7e  confirmation-375.jpg
4e43e4540b811edf68f78d1fb57ee73a94424e060d19e61824f4fda42b194189  executing-375.jpg
98c371664b6283eee2e6684ea9f48dec6d11dc306e3d66e933e04968b4a8ce19  failed-375.jpg
5533bfda969d96ab2146abaf5b79d2dac9611db170c4b5a5785f52d6bcda100c  prequote-1280.jpg
3669fd410354e313d1c3f3d4e61b0567c324ac5f0cfcd04304ad567e0fdb311a  prequote-320.jpg
e9faec724a4b2e11d627b9373d49676c64bb58434c0bcdb3dd11376c36a8d93d  prequote-375.jpg
8d0a3f1e2f7de49b127eda42caf451394844b0224f5b94c4971791413366e1ec  prequote-768.jpg
d156aead5bdf90c92b2a1eac84e3e2f980b549db3643d526dada08cc5dfaf477  quote-review-1280.jpg
fc35c8f4ef49193722793fac76e6e485f83594b9327b070cf32afb6662a8afdd  quote-review-320.jpg
8165fb4c9ca411b01a08e2ef815001dbc2a3ee5b806ce7cf1c3a0a677882dd3f  quote-review-375.jpg
62d14215cc0dafcb8aa2f9cce3f4d48f27b84e0863f9395223bfe0424594a8dd  quote-review-768.jpg
c69805105e121498e8fe01550c5f5a148b9f48fba1802cb39d6043f38e47fccf  recovery-375.jpg
b14e66189411daf9f853a1734f60d6720b8a11d6c81f4f203311f4195e364167  succeeded-375.jpg
```
