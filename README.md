# Object-Diff
This library is designed to provide a mechanism to compare two objects and compute a difference. That difference can then be applied to other objects, performing a selective update to an object.

This library was designed to solve a common problem in Kubernetes operator development, but has been designed to be as general purpose as possible. Bugs and pull requests are welcomed and encouraged.

## Shortcomings
This library is primarily designed to move an object from one state to the next as in A -> B -> C, if trying to enact something like A -> C there is no guarantee that all changes can be captured and could result in partial, possibly incomplete, objects. This should generally be in line with what is expected of a diff and patch algorithm, but can have unexpected consequences.

## Improvements and Future Work
* Currently the tests are lacking and there are possibly cases which are not covered or behave badly. Work on this area is currently in progress.
* Comparision of arrays and slices is currently only by index, future work is planned to provide plugins/callbacks which allow to provide an identity for a particular object in an array or slice and compute the difference based on that.
* For instances when you want to ignore certain known changes, providing a mechanism for exclusions is future work.
* Renaming a map key results in a delete and addition.
* Example usage as part of a Kubernetes operator is currently a near term goal.
