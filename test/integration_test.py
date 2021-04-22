import unittest
import shutil
from os import environ, path
from fmpy import extract
from fmpy.validation import validate_fmu
from fmpy.model_description import read_model_description
from fmpy.simulation import instantiate_fmu, simulateME, simulateCS
import numpy as np

if not 'TEST_FMU' in environ:
    raise(RuntimeError('TEST_FMU environment variable should be set'))
fmu_path = environ['TEST_FMU']

class ValidateFMU(unittest.TestCase):

    def test_fmpy_validate(self):
        errs = validate_fmu(fmu_path)
        self.assertCountEqual(errs, [], 'Errors list should be empty')

class VerifyFMU(unittest.TestCase):

    def test_fmi2GetVersion(self):
        self._skipIfNotCoSim()
        v = self._fmu_cs.getVersion()
        self.assertEqual(self._model_description.fmiVersion, v)

    def test_simulate_model_exchange(self):
        if self._fmu_me is None:
            self.skipTest('Not model exchange')
        simulateME(self._model_description, self._fmu_me, self._start_time, self._stop_time, 
            'CVode', self._step_size, self._relative_tolerance, {}, False, 
            None, None, self._output_interval, True, None, None)

    def test_simulate_co_simulation(self):
        self._skipIfNotCoSim()
        simulateCS(self._model_description, self._fmu_cs, self._start_time, self._stop_time,
            self._relative_tolerance, {}, False, None, None, self._output_interval, None, None)

    def test_state_get_and_set(self):
        # find an FMU with real inputs, get initial state, mutate real value, restore state
        self._skipIfNotCoSim()
        if not self._model_description.coSimulation.canGetAndSetFMUstate:
            self.skipTest('Cannot get and set FMU state')
        setReals = list(filter(lambda v: v.type == 'Real' and v.variability != 'constant' and v.initial == 'exact',
            self._model_description.modelVariables))
        if not setReals:
            self.skipTest('FMU has no settable real variables')
        var = setReals[0]
        self._fmu_cs.setupExperiment(tolerance=self._relative_tolerance, startTime=self._start_time)
        self._fmu_cs.enterInitializationMode()
        v1 = self._fmu_cs.getReal([var.valueReference])[0]
        v2 = v1 + 1.0
        # save initial state
        state = self._fmu_cs.getFMUstate()
        # update variable
        self._fmu_cs.setReal([var.valueReference], [v2])
        self.assertEqual(v2, self._fmu_cs.getReal([var.valueReference])[0])
        # restore original state and check
        self._fmu_cs.setFMUstate(state)
        self.assertEqual(v1, self._fmu_cs.getReal([var.valueReference])[0])
        self._fmu_cs.freeFMUstate(state)

    def test_state_serialize_deserialize(self):
        # find an FMU with real inputs, get initial state, mutate real value, ser/derserialize state and restore state
        self._skipIfNotCoSim()
        if not self._model_description.coSimulation.canSerializeFMUstate:
            self.skipTest('Cannot serialize FMU state')
        setReals = list(filter(lambda v: v.type == 'Real' and v.variability != 'constant' and v.initial == 'exact',
            self._model_description.modelVariables))
        if not setReals:
            self.skipTest('FMU has no settable real variables')
        var = setReals[0]
        self._fmu_cs.setupExperiment(tolerance=self._relative_tolerance, startTime=self._start_time)
        self._fmu_cs.enterInitializationMode()
        v1 = self._fmu_cs.getReal([var.valueReference])[0]
        v2 = v1 + 1.0
        # save initial state
        state = self._fmu_cs.getFMUstate()
        self._fmu_cs.setReal([var.valueReference], [v2])
        self.assertEqual(v2, self._fmu_cs.getReal([var.valueReference])[0])
        # serialize and deserialize state
        ser = self._fmu_cs.serializeFMUstate(state)
        state = self._fmu_cs.deSerializeFMUstate(ser)
        # restore state and verify original value
        self._fmu_cs.setFMUstate(state)
        self.assertEqual(v1, self._fmu_cs.getReal([var.valueReference])[0])
        self._fmu_cs.freeFMUstate(state)

    
    def _skipIfNotCoSim(self):
        if self._fmu_cs is None:
            self.skipTest('Not co simulation')


    @classmethod
    def setUpClass(cls):
        # store various useful FMU arguments from the model description
        cls._model_description = read_model_description(fmu_path)
        cls._fmi_types = []
        if cls._model_description.coSimulation is not None:
            cls._fmi_types.append('CoSimulation')
        if cls._model_description.modelExchange is not None:
            cls._fmi_types.append('ModelExchange')

        if not cls._fmi_types:
            raise Exception('fmi_type must contain at least "ModelExchange" or "CoSimulation"')

        cls._experiment = cls._model_description.defaultExperiment

        if cls._experiment is not None and cls._experiment.startTime is not None:
            cls._start_time = cls._experiment.startTime
        else:
            cls._start_time = 0.0

        start_time = float(cls._start_time)

        if cls._experiment is not None and cls._experiment.stopTime is not None:
            cls._stop_time = cls._experiment.stopTime
        else:
            cls._stop_time = start_time + 1.0

        stop_time = float(cls._stop_time)

        if cls._experiment is not None:
            cls._relative_tolerance = cls._experiment.tolerance

        total_time = stop_time - start_time
        cls._step_size = 10 ** (np.round(np.log10(total_time)) - 3)

        if 'CoSimulation' in cls._fmi_types and cls._experiment is not None and cls._experiment.stepSize is not None:
            cls._output_interval = cls._experiment.stepSize
            while (stop_time - start_time) / cls._output_interval > 1000:
                cls._output_interval *= 2

        if path.isfile(path.join(fmu_path, 'modelDescription.xml')):
            cls._unzipdir = fmu_path
            cls._tempdir = None
        else:
            cls._tempdir = extract(fmu_path)
            cls._unzipdir = cls._tempdir

    @classmethod
    def tearDownClass(cls):
        if cls._tempdir is not None:
            shutil.rmtree(cls._tempdir, ignore_errors=True)

    def setUp(self):
        if 'CoSimulation' in self._fmi_types:
            self._fmu_cs = instantiate_fmu(self._unzipdir, self._model_description, 'CoSimulation', False, True)
        if 'ModelExchange' in self._fmi_types:
            self._fmu_me = instantiate_fmu(self._unzipdir, self._model_description, 'ModelExchange', False, True)

    def tearDown(self):
        if self._fmu_cs is not None:
            self._fmu_cs.freeInstance()
        if self._fmu_me is not None:
            self._fmu_cs.freeInstance()

if __name__ == '__main__':
    unittest.main()