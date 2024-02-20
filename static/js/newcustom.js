$(document).ready(function () {
	// 初始化设置对象
	var settings = JSON.parse(localStorage.getItem('settings')) || {};
	var selectedModelInfo =
		JSON.parse(localStorage.getItem('selectedModelInfo')) || {};

	function initSelect2() {
		$('#model').select2({
			placeholder: '选择模型',
			minimumResultsForSearch: Infinity,
			width: '60%',
			multiple: true,
		});
	}

	function updateProviderSettings(provider) {
		var providerSettings = settings[provider] || {};
		$('#modelEndpoint').val(providerSettings.modelEndpoint || '');
		$('#apiKey').val(providerSettings.apiKey || '');
		updateModelSelect(providerSettings.models || []);
	}

	function updateModelSelect(selectedModels) {
		var defaultModels = [
			{ id: 'gpt-3.5-turbo', text: 'gpt-3.5-turbo' },
			{ id: 'gpt-4', text: 'gpt-4' },
			{ id: 'gpt-4-1106-preview', text: 'gpt-4-1106-preview' },
			{ id: 'gemini-pro', text: 'gemini-pro' },
			{ id: 'qwen-max', text: 'qwen-max' },
			{ id: 'qwen-max-1201', text: 'qwen-max-1201' },
			{ id: 'qwen-max-longcontext', text: 'qwen-max-longcontext' },
		];

		var modelOptions = defaultModels.map(function (model) {
			return new Option(
				model.text,
				model.id,
				selectedModels.includes(model.id),
				selectedModels.includes(model.id)
			);
		});

		$('#model').empty().append(modelOptions).trigger('change');
	}

	function saveProviderSettings(provider) {
		settings[provider] = {
			modelEndpoint: $('#modelEndpoint').val(),
			apiKey: $('#apiKey').val(),
			models: $('#model').val() || [],
		};
		localStorage.setItem('settings', JSON.stringify(settings));
	}

	function updateSingleModelOptions() {
		var $singleModelSelect = $('#singleModel');
		$singleModelSelect.empty();

		$singleModelSelect.append(
			$('<option>', {
				value: '{"modelId":"gpt-3.5-turbo","provider":"fucker"}',
				text: '内置免费🔖gpt3.5',
			})
		);

		Object.keys(settings).forEach(function (provider) {
			(settings[provider].models || []).forEach(function (modelId) {
				var modelValue = JSON.stringify({
					modelId: modelId,
					provider: provider,
				});
				$singleModelSelect.append(
					new Option(modelId + '🏷️' + provider, modelValue)
				);
			});
		});

		setSelectedModel();
	}

	function setSelectedModel() {
		var $singleModelSelect = $('#singleModel');
		var isValid =
			settings[selectedModelInfo.provider] &&
			settings[selectedModelInfo.provider].models.includes(
				selectedModelInfo.modelId
			);

		// 使用JSON.stringify确保option的value和selectedModelInfo匹配
		var selectedModelValue = JSON.stringify({
			modelId: selectedModelInfo.modelId,
			provider: selectedModelInfo.provider,
		});

		if (isValid || selectedModelInfo.provider === 'fucker') {
			$singleModelSelect.val(selectedModelValue).trigger('change');
		} else {
			$singleModelSelect
				.val('{"modelId":"gpt-3.5-turbo","provider":"fucker"}')
				.trigger('change');
			storeSelectedModelInfo('{"modelId":"gpt-3.5-turbo","provider":"fucker"}');
		}
	}

	function storeSelectedModelInfo(selectedModelValue) {
		try {
			var selectedModel = JSON.parse(selectedModelValue);
			if (selectedModel && selectedModel.provider && selectedModel.modelId) {
				selectedModelInfo = {
					modelId: selectedModel.modelId,
					provider: selectedModel.provider,
					modelEndpoint: settings[selectedModel.provider]?.modelEndpoint,
					apiKey: settings[selectedModel.provider]?.apiKey,
				};
				localStorage.setItem(
					'selectedModelInfo',
					JSON.stringify(selectedModelInfo)
				);
			}
		} catch (error) {
			console.error('Error parsing selected model value:', error);
		}
	}

	function restoreSelectedModelInfo() {
		var storedModelInfo = localStorage.getItem('selectedModelInfo');
		if (storedModelInfo) {
			try {
				selectedModelInfo = JSON.parse(storedModelInfo);
				updateSingleModelOptions(); // 确保选项已经填充
				setSelectedModel(); // 然后设置选中的模型
			} catch (error) {
				console.error('Error parsing stored selected model info:', error);
				$('#singleModel').val('builtin').trigger('change');
			}
		} else {
			$('#singleModel').val('builtin').trigger('change');
		}
	}

	$('#serviceProvider').change(function () {
		var provider = $(this).val();
		updateProviderSettings(provider);
	});

	$('#singleModel').change(function () {
		storeSelectedModelInfo($(this).val());
	});

	$('#modelEndpoint, #apiKey, #model').on('change', function () {
		var provider = $('#serviceProvider').val();
		saveProviderSettings(provider);
		updateSingleModelOptions();
	});

	initSelect2();
	var defaultProvider = $('#serviceProvider').val();
	updateProviderSettings(defaultProvider); // 加载服务商设置
	restoreSelectedModelInfo(); // 恢复选中的模型信息
});
