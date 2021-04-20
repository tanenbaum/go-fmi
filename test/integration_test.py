import unittest
from fmpy.validation import validate_fmu

class ValidateFMU(unittest.TestCase):

    def test_validate(self):
        errs = validate_fmu('./out/fmus/BouncingBall.fmu')
        self.assertCountEqual(errs, [], 'Errors list should be empty')

if __name__ == '__main__':
    unittest.main()